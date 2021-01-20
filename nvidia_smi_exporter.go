package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "net/http"
    "log"
    "os"
    "os/exec"
    "strings"
)


func dmon(response http.ResponseWriter, request *http.Request) {
    out, err := exec.Command("nvidia-smi", "dmon", "-c", "1", "-s", "u").Output()

    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }

    reader := bufio.NewReader(bytes.NewReader(out))
    line, isPrefix, err := reader.ReadLine()
    line, isPrefix, err = reader.ReadLine()
    line, isPrefix, err = reader.ReadLine()
    var records [][]string
    for err == nil && !isPrefix {
        scanner := bufio.NewScanner(strings.NewReader(string(line)))
        scanner.Split(bufio.ScanWords)
        count := 0
        var row []string
        for scanner.Scan() && count < 5 {
            row = append(row, scanner.Text())
        }
        records = append(records, row)
        line, isPrefix, err = reader.ReadLine()
    }

    metricList := []string {
        "utilization.sm", "utilization.mem", "utilization.enc", "utilization.dec"}

    result := ""
    for _, row := range records {
        name := fmt.Sprintf("GPU[%s]", row[0])
        for idx, value := range row[1:] {
            result = fmt.Sprintf(
                "%s%s{gpu=\"%s\"} %s\n", result,
                metricList[idx], name, value)
        }
    }

    fmt.Fprintf(response, strings.Replace(result, ".", "_", -1))
}

// name, index, temperature.gpu, utilization.gpu,
// utilization.memory, memory.total, memory.free, memory.used
// clocks.gr, clocks.video, clocks.sm

func metrics(response http.ResponseWriter, request *http.Request) {
    out, err := exec.Command(
        "nvidia-smi",
        "--query-gpu=name,index,temperature.gpu,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,clocks.gr,clocks.video,clocks.sm",
        "--format=csv,noheader,nounits").Output()

    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }

    csvReader := csv.NewReader(bytes.NewReader(out))
    csvReader.TrimLeadingSpace = true
    records, err := csvReader.ReadAll()

    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }

    metricList := []string {
        "temperature.gpu", "utilization.gpu",
        "utilization.memory", "memory.total", "memory.free", "memory.used"}

    result := ""
    for _, row := range records {
        name := fmt.Sprintf("%s[%s]", row[0], row[1])
        for idx, value := range row[2:] {
            result = fmt.Sprintf(
                "%s%s{gpu=\"%s\"} %s\n", result,
                metricList[idx], name, value)
        }
    }

    fmt.Fprintf(response, strings.Replace(result, ".", "_", -1))
}

func main() {
    addr := ":9101"
    if len(os.Args) > 1 {
        addr = ":" + os.Args[1]
    }

    http.HandleFunc("/metrics/", metrics)
    http.HandleFunc("/dmon/", dmon)
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
