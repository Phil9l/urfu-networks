package main

import (
    "log"
    "net"
    "fmt"
    "time"
)

const PORT = 8080
const DELAY = 1000 // In Millisecond

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
            log.Println("New connection accepted.")
            go handleConnection(conn)
    }
}

func handleConnection(connection net.Conn) {
    connection.Write([]byte("Hi!\n"))
    time.Sleep(DELAY * time.Millisecond)
    connection.Write([]byte("Bye!\n"))
    connection.Close()
}
