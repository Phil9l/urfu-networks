package main

import (
    "bufio"
    "log"
    "net"
    "fmt"
    "time"
)

const PORT = 8080
const DELAY = 0 // In Millisecond

func main() {
    log.Printf("Starting echo-server on port %d", PORT)

    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
    if err != nil {
        log.Fatalf("Listen error: %s", err)
    }

    for {
            conn, err := ln.Accept()
            if err != nil {
                log.Fatalf("Accept error: %s", err)
            }
            go handleConnection(conn)
            log.Println("New connection accepted.")
    }
}

func handleConnection(connection net.Conn) {
    buffer := bufio.NewReader(connection)

    for {
        line, err := buffer.ReadBytes('\n')
        if err != nil {
            break
        }
        log.Printf("Received new message: %s", line)
        time.Sleep(DELAY * time.Millisecond)
        connection.Write(line)
    }
}
