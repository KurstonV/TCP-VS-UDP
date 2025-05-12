package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	clients = make(map[string]bool) // Track clients by address string (IP:port)
	mutex   = sync.Mutex{}
)

const (
	serverAddr = "localhost:8080"
)

type Client struct {
	addr *net.UDPAddr
    lastSeen time.Time
}

func main() {
	// Resolve the server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Println("Address resolution error:", err)
		return
	}

	// Listen for UDP connections
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP Server started on :8080")

	buffer := make([]byte, 1024) // large enough to test overflow

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}

		go handleUDPMessage(conn, clientAddr, buffer[:n])
	}
}

func handleUDPMessage(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	message := strings.TrimSpace(string(data))
	clientID := addr.String()

	mutex.Lock()
	if _, exists := clients[clientID]; !exists {
		clients[clientID] = true
		// Notify new client of connected clients
		conn.WriteToUDP([]byte(fmt.Sprintf("Welcome! %d client(s) connected.\n", len(clients))), addr)
	}
	mutex.Unlock()

    

	// Create a log file named after client address (e.g., 127.0.0.1_49230.log)
	logFileName := strings.ReplaceAll(clientID, ":", "_") + ".log"
	logFile, _ := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	if len(message) == 0 {
		conn.WriteToUDP([]byte("Say something..."), addr)
		return
	}

	if len(message) > 1024 {
		conn.WriteToUDP([]byte("Message too long. Max 1024 bytes."), addr)
		return
	}

	// Log message to client's log file
	logFile.WriteString(message + "\n")

	// Command and personality logic
	switch {
	case message == "/time":
		conn.WriteToUDP([]byte(time.Now().Format(time.RFC1123)), addr)
	case message == "/quit":
		conn.WriteToUDP([]byte("Disconnecting..."), addr)
		mutex.Lock()
		delete(clients, clientID)
		mutex.Unlock()
	case strings.HasPrefix(message, "/echo "):
		echoMsg := strings.TrimPrefix(message, "/echo ")
		conn.WriteToUDP([]byte(echoMsg), addr)
	default:
		// Broadcast message to all other clients
		mutex.Lock()
		for id := range clients {
			if id != clientID {
				targetAddr, _ := net.ResolveUDPAddr("udp", id)
				conn.WriteToUDP([]byte(fmt.Sprintf("[%s]: %s\n", clientID, message)), targetAddr)
			}
		}
		mutex.Unlock()
	}
}