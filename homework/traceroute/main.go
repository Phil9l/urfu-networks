package main

import (
    "fmt"
    "net"
    "syscall"
    "flag"
    "log"
)

const (
    HOST = "0.0.0.0"
    RECV_PORT = 0
    MAX_TTL = 32
)


func main() {
	flag.Usage = func() {
		fmt.Printf("Usage:\n./main [options] hostname\nOptions:\n")
		flag.PrintDefaults()
	}
    flag.Parse()
    if flag.NArg() != 1 {
        flag.Usage()
		return
    }

    address, err := net.LookupHost(flag.Args()[0])
    if err != nil {
        log.Println("Bad hostname")
    }
    traceroute(address[0])
}

func traceroute(address string) {
    log.Printf("Starting")
    for ttl := 1; ttl <= MAX_TTL; ttl++ {
        resIP, err := tracerouteWithGivenTTL(getIPv4Address(address, 35353), ttl)
        // log.Printf(resIP)
        if err != nil {
            fmt.Printf("%d\t* * *\n", ttl)
        } else {
            netName, origin, country, err := whois(resIP)
            if err == nil {
                fmt.Printf("%d\t%s\t[%s\t%s\t%s]\n", ttl, resIP, netName, origin, country)
                // fmt.Printf("%s %s [⚑ %s]\n", netName, origin, country)
            } else {
                fmt.Printf("%d\t%s\n", ttl, resIP)
            }
        }
    }
}

func tracerouteWithGivenTTL(address *syscall.SockaddrInet4, ttl int) (string, error) {
    // log.Printf("Tracerouting %s", address)
    writeSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

    if err != nil {
        return "", err
    }

    readSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
    if err != nil {
        return "", err
    }

    defer syscall.Close(readSock)
    defer syscall.Close(writeSock)
    
    tv := syscall.NsecToTimeval(1e6 * 5000)
    err = syscall.SetsockoptTimeval(readSock, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
    if err != nil {
		return "", err
	}

    err = syscall.SetsockoptInt(writeSock, syscall.SOL_IP, syscall.IP_TTL, ttl)
    if err != nil {
        return "", err
    }
    
    err = syscall.Bind(readSock, getIPv4Address(HOST, RECV_PORT))
    if err != nil {
        return "", err
    }
    
    icmp_data := []byte{8, 0, 247, 255, 0, 0, 0, 0}    
    err = syscall.Sendto(writeSock, icmp_data, 0, address)
    if err != nil {
        return "", err
    }

    buf := make([]byte, 512)
	_, from, err := syscall.Recvfrom(readSock, buf, 0)
	if err != nil {
		return "", err
	}
    
    // err_type, err_code := int(buf[20]), int(buf[21])
    
    tmp := from.(*syscall.SockaddrInet4).Addr
    // log.Printf("Error: [%d, %d]\n", err_type, err_code)
    // log.Printf("%d\t%d.%d.%d.%d\n", ttl, tmp[0], tmp[1], tmp[2], tmp[3])
    
    return fmt.Sprintf("%d.%d.%d.%d", tmp[0], tmp[1], tmp[2], tmp[3]), nil
}

func getIPv4Address(hostname string, port int) *syscall.SockaddrInet4 {
    ipArray := net.ParseIP(hostname).To4()
    return &syscall.SockaddrInet4 {
        Port: port,
        Addr: [4]byte{ipArray[0], ipArray[1], ipArray[2], ipArray[3]},
    }
}
