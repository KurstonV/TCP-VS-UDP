package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    localAddr, _ := net.ResolveUDPAddr("udp", ":0")
    serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:9001")

    conn, _ := net.DialUDP("udp", localAddr, serverAddr)
    defer conn.Close()

    go func() {
        buf := make([]byte, 1024)
        for {
            n, _, err := conn.ReadFromUDP(buf)
            if err != nil {
                return
            }
            fmt.Print(string(buf[:n]))
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        conn.Write([]byte(scanner.Text()))
    }
}
