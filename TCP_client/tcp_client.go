package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:9000")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        fmt.Fprintln(conn, scanner.Text())
    }
}
