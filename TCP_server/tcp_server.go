package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "sync"
    "time"
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
     addr := conn.RemoteAddr().String()
    logFileName := strings.ReplaceAll(addr, ":", "_") + ".log"
    logFile, _ := os.Create(logFileName) // logs overwritten per run
    defer logFile.Close()
    
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

     for {
        // ⏱ Set 30-second timeout for reading input
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))

        if !scanner.Scan() {
            err := scanner.Err()
            if err != nil {
                // Handle timeout separately
                netErr, ok := err.(net.Error)
                if ok && netErr.Timeout() {
                    fmt.Fprintln(conn, "Disconnected due to 30 seconds of inactivity.")
                }
            }
            return
        }

        input := scanner.Text()

        if len(input) == 0 {
            fmt.Fprintln(conn, "Say something...")
            continue
        }

        if len(input) > 1024 {
            fmt.Fprintln(conn, "Message too long. Max 1024 bytes.")
            continue
        }
//write to the logfile
        logFile.WriteString(input + "\n")
 //Response Protocols
        switch {
        case input == "/time":
            fmt.Fprintln(conn, "Server time:", time.Now().Format(time.RFC1123))
        case input == "/quit":
            fmt.Fprintln(conn, "Disconnecting...")
            return
        case strings.HasPrefix(input, "/echo "):
            fmt.Fprintln(conn, strings.TrimPrefix(input, "/echo "))
        default:
            msg := fmt.Sprintf("[%s]: %s\n", conn.RemoteAddr(), input)
            broadcast(msg, conn)
        }
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
