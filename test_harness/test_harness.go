package main

import (
    "bufio"
    "fmt"
    "math/rand"
    "net"
    "strings"
    "sync"
    "time"
)

const (
    serverAddr        = "localhost:9000"
    numClients        = 10
    messagesPerClient = 20
)

type Metrics struct {
    Latency      []time.Duration
    BytesSent    int
    MessagesSent int
    MessagesRecv int
    sync.Mutex
}

func main() {
    var wg sync.WaitGroup
    metrics := &Metrics{}

    for i := 0; i < numClients; i++ {
        wg.Add(1)
        go simulateClient(i, metrics, &wg)
        time.Sleep(100 * time.Millisecond) // stagger client joins
    }

    wg.Wait()
    report(metrics)
}

func simulateClient(id int, metrics *Metrics, wg *sync.WaitGroup) {
    defer wg.Done()

    conn, err := net.Dial("tcp", serverAddr)
    if err != nil {
        fmt.Printf("Client %d failed to connect: %v\n", id, err)
        return
    }
    defer conn.Close()

    received := 0
    sent := 0
    seqPrefix := fmt.Sprintf("Client %d:", id)

    // Receiver goroutine: Count matching echoes
    done := make(chan struct{})
    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            line := scanner.Text()
            if strings.Contains(line, seqPrefix) {
                received++
            }
        }
        close(done)
    }()

    for i := 0; i < messagesPerClient; i++ {
        msg := fmt.Sprintf("%s msg %d", seqPrefix, i)
        start := time.Now()
        n, err := fmt.Fprintln(conn, msg)
        if err != nil {
            fmt.Printf("Client %d write error: %v\n", id, err)
            return
        }

        latency := time.Since(start)

        metrics.Lock()
        metrics.Latency = append(metrics.Latency, latency)
        metrics.BytesSent += n
        metrics.MessagesSent++
        metrics.Unlock()

        sent++

        jitter := time.Duration(rand.Intn(200)) * time.Millisecond
        time.Sleep(300*time.Millisecond + jitter)

        if rand.Float32() < 0.05 {
            fmt.Printf("Client %d disconnecting early\n", id)
            return
        }
    }

    time.Sleep(1 * time.Second) // wait for final messages
    <-done

    metrics.Lock()
    metrics.MessagesRecv += received
    metrics.Unlock()
}

func report(metrics *Metrics) {
    metrics.Lock()
    defer metrics.Unlock()

    var totalLatency time.Duration
    for _, l := range metrics.Latency {
        totalLatency += l
    }

    avgLatency := time.Duration(0)
    if len(metrics.Latency) > 0 {
        avgLatency = totalLatency / time.Duration(len(metrics.Latency))
    }

    loss := 0
    if metrics.MessagesSent > metrics.MessagesRecv {
        loss = metrics.MessagesSent - metrics.MessagesRecv
    }

    lossPercent := float64(loss) / float64(metrics.MessagesSent) * 100

    fmt.Println("====== Test Report ======")
    fmt.Printf("Clients: %d\n", numClients)
    fmt.Printf("Messages Sent: %d\n", metrics.MessagesSent)
    fmt.Printf("Messages Received (echo): %d\n", metrics.MessagesRecv)
    fmt.Printf("Packet Loss: %.2f%%\n", lossPercent)
    fmt.Printf("Avg Latency: %v\n", avgLatency)
    fmt.Printf("Total Bytes Sent: %d\n", metrics.BytesSent)
    fmt.Printf("Throughput: %.2f KB\n", float64(metrics.BytesSent)/1024)
    fmt.Println("==========================")
}
