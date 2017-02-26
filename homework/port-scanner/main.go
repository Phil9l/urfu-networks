package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
    "os"
    "bufio"
)

const DELAY = 1 // In seconds.
const MIN_PORT = 0
const MAX_PORT = 65535

func main() {
    if len(os.Args) != 2 {
        log.Fatal("Required 1 argument: IP to scan")
    }
    hostnameToScan := os.Args[1]

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
	connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), DELAY*time.Second)
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
    wg.Add(1)

    go checkSSHprotocol(address, protocolChannel, wg)

    wg.Wait()
    
    protocol := "Unknown"
    select {
        case x, ok := <-protocolChannel:
            log.Println(address)
            if ok {
                protocol = x
            }
        default:
    }
    resultChannel <- struct {address, status string}{address, protocol} 
}

func checkSSHprotocol(address string, responseChannel chan string, wg *sync.WaitGroup) {
    connection, err := net.Dial("tcp", address)
    connection.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
    bufReader := bufio.NewReader(connection)
    result, err := bufReader.ReadBytes('\n')

    if err == nil && strings.HasPrefix(string(result), "SSH") {
        responseChannel <- "SSH"
    }

    wg.Done()
}
