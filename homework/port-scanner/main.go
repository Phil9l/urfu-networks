package main

import (
    "log"
    "net"
    "fmt"
    "time"
    "bufio"
    "strings"
)

const DELAY = 1  // In seconds.
const MIN_PORT = 0
const MAX_PORT = 65535

func main() {
    // log.Println(IsTCPPortAvailable("127.0.0.1", 80));
    // log.Println(IsTCPPortAvailable("127.0.0.1", 81));
    for i := MIN_PORT; i < MAX_PORT; i++ {
        var isOpen, connection = IsTCPPortAvailable("188.225.77.148", i)
        // log.Printf("%d — %t\n", i, isOpen)
        if isOpen {
            var protocol, _ = CheckProtocol(connection)
            log.Printf("%d — %s\n", i, protocol)
        }
    }
}

func IsTCPPortAvailable(hostname string, port int) (bool, net.Conn) {
    connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), DELAY * time.Second)
    if err == nil {
        return true, connection
    } else {
        return false, nil
    }
}

func CheckProtocol(connection net.Conn) (string, error) {
    connection.Write([]byte{0})

    buffer := bufio.NewReader(connection)

    result, err := buffer.ReadBytes('\n')
    connection.Close()
    if err != nil {
        return "Unknown", err
    }

    if strings.HasPrefix(string(result), "SSH") {
        return "SSH", nil
    }
    if strings.HasPrefix(string(result), "HTTP") {
        return "HTTP", nil
    }

    return "Unknown", nil
}
