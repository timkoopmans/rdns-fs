package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "regexp"
    "sync"
    "github.com/cheggaaa/pb"
    "net"
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

    filePath := flag.String("file", "test.json", "file path to read from")
    workers := flag.Int("workers", 500, "number of concurrent workers")
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

    for i := 0; i < *workers; i++ {
        wg.Add(1)
        go processRows(rows, bar, cidrs, &wg)
    }

    wg.Wait()
    bar.Finish()
}

func processRows(rows <-chan string, bar *pb.ProgressBar, cidrs []string, wg *sync.WaitGroup) {
    m := regexp.MustCompile(`\.`)
    
    for row := range rows {
        var record Record
        json.Unmarshal([]byte(row), &record)
        ip := record.Name

        if checkFirstOctet(ip) {
            if checkSubnetsContainAddress(ip, cidrs) {
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

                bytes, err := outfile.WriteString(record.Value)
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

func checkFirstOctet(ip string) bool {
    filter := [100]string{
        "103","108","116","118","119","120","13","130","140","143","144","15","150","161","172","176","177","178",
        "18","180","185","199","203","204","205","207","209","216","223","27","3","34","35","36","43","44","52","54",
        "58","63","64","65","69","70","71","72","75","76","87","99",
    }

    firstOctet := strings.Split(ip, ".")

    for _, v := range filter {
        if v == firstOctet[0] {
            return true
        }
    }

    return false
}

func checkSubnetsContainAddress(ip string, cidrs []string) bool {
    address := net.ParseIP(ip)

    for _, cidr := range cidrs {
        _, subnet, _ := net.ParseCIDR(cidr)

        if subnet.Contains(address) {
            return true
        }
    }

    return false
}
