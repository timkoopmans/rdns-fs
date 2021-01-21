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
    "sync"
    "strings"
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
    lines := make(chan string)
    done := make(chan bool)

    filePath := flag.String("file", "test.json", "file path to read IPs from")
    cidrsPath := flag.String("cidr", "cidrs.txt", "file path to read CIDRs from")
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

    cidrs, err := readLines(*cidrsPath)
    if err != nil {
        fmt.Println("unable to read lines from the file", err)
    }

    ranger := cidranger.NewPCTrieRanger()

    for _, cidr := range cidrs {
        _, network, _ := net.ParseCIDR(cidr)
        ranger.Insert(cidranger.NewBasicRangerEntry(*network))
    }

    for i := 0; i < *workers; i++ {
        wg.Add(1)
        go filterRows(rows, lines, bar, ranger, &wg)
    }

    go writeLines(lines, done)

    wg.Wait()
    close(lines)

    d := <-done
    if d == true {
        bar.Finish()
    } else {
        fmt.Println("File writing failed")
    }
}

func filterRows(rows <-chan string, lines chan<- string, bar *pb.ProgressBar, ranger cidranger.Ranger, wg *sync.WaitGroup) {
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
            lines <- row
            bar.Add(len(row))
        } else {
            bar.Add(len(row) + 1)
        }
    }

    wg.Done()
}

func writeLines(lines chan string, done chan bool) {
    for line := range lines {
        var record Record
        json.Unmarshal([]byte(line), &record)

        ip := record.Name
        firstOctet := strings.Split(ip, ".")[0]

        filename := fmt.Sprintf("rdns.%s.0.0.0.json", firstOctet)

        file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println("unable to open the file", err)
            return
        }

        _, err = fmt.Fprintln(file, line)
        if err != nil {
            fmt.Println("unable to write the file", err)
            file.Close()
            return
        }

        err = file.Close()
        if err != nil {
            fmt.Println("unable to close the file", err)
            return
        }

        done <- true
    }
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
