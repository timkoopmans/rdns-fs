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

func processRow(rows <-chan string, bar *pb.ProgressBar) {
    for row := range rows {
        var record Record
        json.Unmarshal([]byte(row), &record)
        m := regexp.MustCompile(`\.`)
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
}

func processFile(file *os.File, workers *int) {
    var wg sync.WaitGroup

    filestat, err := file.Stat()
    if err != nil {
        fmt.Println("unable to get file stat")
        return
    }

    fileSize := filestat.Size()

    bar := pb.StartNew(int(fileSize))

    rows := make(chan string)

    for w := 1; w <= *workers; w++ {
        wg.Add(1)
        go func() {
            processRow(rows, bar)
            wg.Done()
        }()
    }

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
}

func main() {
    var wg sync.WaitGroup

    fptr := flag.String("file", "test.json", "file path to read from")
    workers := flag.Int("workers", 500, "number of concurrent workers")
    flag.Parse()

    file, err := os.Open(*fptr)
    if err != nil {
        fmt.Println("unable to open the file", err)
        return
    }
    defer file.Close()

    wg.Add(1)
    go processFile(file, workers)
    wg.Wait()
}
