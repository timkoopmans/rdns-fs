package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/cheggaaa/pb"
    "github.com/yl2chen/cidranger"
    "net"
    "os"
    "regexp"
    "sync"
)

type Record struct {
    Timestamp string
    Name string
    Value string
    Type string
}

func main() {
    var wg sync.WaitGroup

    rows := make(chan string)

    filePath := flag.String("file", "test.json", "file path to read from")
    workers := flag.Int("workers", 50, "number of concurrent workers")
    flag.Parse()

    file, err := os.Open(*filePath)
    if err != nil {
        fmt.Println("unable to open the file", err)
        return
    }
    defer file.Close()

    fileStat, err := file.Stat()
    if err != nil {
        fmt.Println("unable to get file stat")
        return
    }

    fileSize := fileStat.Size()

    bar := pb.StartNew(int(fileSize))

    go func() {
        scanner := bufio.NewScanner(file)

        for scanner.Scan() {
            rows <- scanner.Text()
        }
        if err := scanner.Err(); err != nil {
            fmt.Println("unable to scan the file", err)
        }
        close(rows)
    }()

    cidrs, err := readLines("cidrs.txt")
    if err != nil {
        fmt.Println("unable to readlines from file", err)
    }

    ranger := cidranger.NewPCTrieRanger()

    for _, cidr := range cidrs {
        _, network, _ := net.ParseCIDR(cidr)
        ranger.Insert(cidranger.NewBasicRangerEntry(*network))
    }

    for i := 0; i < *workers; i++ {
        wg.Add(1)
        go processRows(rows, bar, ranger, &wg)
    }

    wg.Wait()
    bar.Finish()
}

func processRows(rows <-chan string, bar *pb.ProgressBar, ranger cidranger.Ranger, wg *sync.WaitGroup) {
    m := regexp.MustCompile(`\.`)

    for row := range rows {
        var record Record
        json.Unmarshal([]byte(row), &record)
        ip := record.Name

        contains, err := ranger.Contains(net.ParseIP(ip))
        if err != nil {
            fmt.Println("unable to parse IP", err)
            return
        }
        if contains {
            res := m.ReplaceAllString(record.Name, "/")
            path := fmt.Sprintf("rdns/%s", res)
            filename := fmt.Sprintf("rdns/%s/%s", res, record.Timestamp)

            if _, err := os.Stat(filename); os.IsNotExist(err) {
                os.MkdirAll(path, 0700)
            }

            outfile, err := os.Create(filename)

            if err != nil {
                fmt.Println("unable to create the file", err)
                return
            }

            bytes, err := outfile.WriteString(record.Value + "\n")
            if err != nil {
                fmt.Println("unable to write the file", bytes, err)
                outfile.Close()
                return
            }
            bar.Add(len(row))

            err = outfile.Close()
            if err != nil {
                fmt.Println("some other error writing the file", err)
                return
            }
        } else {
            bar.Add(len(row) + 1)
        }
    }

    wg.Done()
}

func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}
