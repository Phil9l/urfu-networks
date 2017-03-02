package main

import (
    "fmt"
    // "log"
    "net"
    "strings"
    "sync"
    "time"
    "bufio"
    "flag"
)

const MIN_PORT = 0
const MAX_PORT = 65535

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage:\n./main [options] ip\nOptions:\n")
		flag.PrintDefaults()
	}
    flag.Parse()

    if flag.NArg() != 1 {
        flag.Usage()
		return
    }
    hostnameToScan := flag.Args()[0]

    portChannel := make(chan int, 100)
    addressChannel := make(chan string, 100)
    resultChannel := make(chan struct {address, status string}, 100)

    go addPortsToChannel(portChannel)
    go selectOpenPorts(hostnameToScan, addressChannel, portChannel)
    go getPortInformation(addressChannel, resultChannel)

    done := make(chan interface{}, 1)

    fmt.Printf("PORT\tSERVICE\n")
    for i := 0; i < 10; i++ {
        go printConnectionInfo(resultChannel, done)
    }
    <-done
}

func printConnectionInfo(ch <-chan struct {address, status string}, done chan interface{}) {
    defer close(done)
    for address := range ch {
        fmt.Printf("%s\t%s\n", strings.Split(address.address, ":")[1], address.status)
    }
}

func addPortsToChannel(ch chan<- int) {
    defer close(ch)
    for i := MIN_PORT; i < MAX_PORT; i++ {
        ch <- i
    }
}

func selectOpenPorts(hostname string, ch chan string, chanPorts <-chan int) {
    wg := &sync.WaitGroup{}
    routineCount := 1000
    wg.Add(routineCount)

    for i := 0; i < routineCount; i++ {
        go func() {
            for i := range chanPorts {
                handleAvailableTCPPorts(hostname, i, ch)
            }
            wg.Done()
        }()
    }
    wg.Wait()
    close(ch)
}

func getPortInformation(addressChannel <-chan string, resultChannel chan struct {address, status string}) {
    for address := range addressChannel {
        go checkProtocol(address, resultChannel)
    }
    close(resultChannel)
}

func handleAvailableTCPPorts(hostname string, port int, ch chan string) {
    connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), 1000 * time.Millisecond)
    if err == nil {
        address := connection.RemoteAddr().String()
        connection.Close()
        ch <- address
    } else {
        // log.Println(err)
    }
}

func checkProtocol(address string, resultChannel chan struct {address, status string}) {
    protocolChannel := make(chan string, 100)
    wg := &sync.WaitGroup{}
    wg.Add(3)
    
    go checkSSHprotocol(address, protocolChannel, wg)
    go checkHTTPprotocol(address, protocolChannel, wg)
    go checkSMTPprotocol(address, protocolChannel, wg)

    wg.Wait()
    
    protocol := "Unknown"
    select {
        case x, ok := <-protocolChannel:
            if ok {
                protocol = x
            }
        default:
    }
    resultChannel <- struct {address, status string}{address, protocol} 
}

func checkSSHprotocol(address string, responseChannel chan string, wg *sync.WaitGroup) {
    connection, err := net.Dial("tcp", address)
    connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
    bufReader := bufio.NewReader(connection)
    result, err := bufReader.ReadBytes('\n')

    if err == nil && strings.HasPrefix(string(result), "SSH") {
        responseChannel <- "SSH"
    }

    wg.Done()
}

func checkHTTPprotocol(address string, responseChannel chan string, wg *sync.WaitGroup) {
    connection, err := net.Dial("tcp", address)
    connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
    connection.Write([]byte("GET / HTTP/1.1\n\n"))
    bufReader := bufio.NewReader(connection)
    result, err := bufReader.ReadBytes('\n')

    if err == nil && strings.HasPrefix(string(result), "HTTP") {
        // log.Println(address, " — HTTP")
        responseChannel <- "HTTP"
    }

    wg.Done()
}

func checkSMTPprotocol(address string, responseChannel chan string, wg *sync.WaitGroup) {
    // log.Println(address)
    connection, err := net.Dial("tcp", address)
    connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
    bufReader := bufio.NewReader(connection)

    _, _ = bufReader.ReadBytes('\n')
    // log.Println(tres)
    connection.Write([]byte("HELO"))
    result, err := bufReader.ReadBytes('\n')
    
    // log.Println(result)

    if err == nil && strings.HasPrefix(string(result), "501") {
        responseChannel <- "SMTP"
    }

    wg.Done()
}
