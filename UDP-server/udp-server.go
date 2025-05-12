package main

import (
    "fmt"
    "net"
)

var clients = make(map[string]*net.UDPAddr)

func main() {
    addr, _ := net.ResolveUDPAddr("udp", ":9001")
    conn, _ := net.ListenUDP("udp", addr)
    defer conn.Close()

    buf := make([]byte, 1024)
    for {
        n, clientAddr, _ := conn.ReadFromUDP(buf)
        msg := string(buf[:n])

        if _, exists := clients[clientAddr.String()]; !exists {
            clients[clientAddr.String()] = clientAddr
        }

        for _, addr := range clients {
            if addr.String() != clientAddr.String() {
                conn.WriteToUDP([]byte(fmt.Sprintf("[%s]: %s", clientAddr.String(), msg)), addr)
            }
        }
    }
}
