package main

import (
    "bufio"
    "fmt"
    "net"
    "sync"
)

 // A map to store connected clients by their connection object
var (
    clients = make(map[net.Conn]string)
    mutex   = sync.Mutex{}
)


func main() {
     // Start listening on TCP port 9000
    ln, err := net.Listen("tcp", ":9000")
    if err != nil {
        panic(err)
    }
    fmt.Println("TCP Server started on :9000")

      // Accept clients in a loop
    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        // Handle each client in its own goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    // Ensure cleanup happens on disconnect
    defer func() {
        mutex.Lock()
        delete(clients, conn)
        mutex.Unlock()
        conn.Close()
        // Notify others of disconnection
         broadcast(fmt.Sprintf("%s has disconnected\n", conn.RemoteAddr()), conn)
    }()
    // Add the new client to the map
    mutex.Lock()
    clients[conn] = conn.RemoteAddr().String()

     // Send connected client count to the new client
    count := len(clients)
    fmt.Fprintf(conn, "Welcome! There are %d client(s) connected.\n", count)

    mutex.Unlock()

    broadcast(fmt.Sprintf("%s joined\n", conn.RemoteAddr()), conn)

    // Create a scanner to read lines from the client
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        // Format message with sender's address
        msg := fmt.Sprintf("[%s]: %s\n", conn.RemoteAddr(), scanner.Text())
        // Send message to all other clients
        broadcast(msg, conn)
    }
}

func broadcast(message string, sender net.Conn) {
    mutex.Lock()
    defer mutex.Unlock()
    
    // Send the message to all connected clients except the sender
    for conn := range clients {
        if conn != sender {
            fmt.Fprint(conn, message)
        }
    }
}
