
// client.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	serverAddr = "localhost:8080"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter your name: ")
    clientName, _ := reader.ReadString('\n')
    clientName = strings.TrimSpace(clientName)

    // Resolve the server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Println("Address resolution error:", err)
		return
	}

	// Connect to the UDP server
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()

    registerMessage := fmt.Sprintf("register:%s", clientName)
    _, err = conn.Write([]byte(registerMessage))
    if err != nil {
        fmt.Println("Failed to register client:", err)
        return
    }
    fmt.Println("Registered with server as:", clientName)
    
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

