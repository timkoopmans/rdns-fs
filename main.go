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

    for i := 0; i < *workers; i++ {
        wg.Add(1)
        go processRows(rows, bar, &wg)
    }

    wg.Wait()
    bar.Finish()
}

func processRows(rows <-chan string, bar *pb.ProgressBar, wg *sync.WaitGroup) {
    m := regexp.MustCompile(`\.`)
    
    for row := range rows {
        var record Record
        json.Unmarshal([]byte(row), &record)
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
        l, err := outfile.WriteString(row)
        if err != nil {
            fmt.Println("unable to write the file", err)
            outfile.Close()
            return
        }
        bar.Add(l)
        err = outfile.Close()
        if err != nil {
            fmt.Println("some other error writing the file", err)
            return
        }
    }

    wg.Done()
}

