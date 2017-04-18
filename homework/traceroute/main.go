package main

import (
    "fmt"
    "net"
    "syscall"
    "flag"
    "log"
    "errors"
)

const (
    RECV_PORT = 33434
    MAX_TTL = 30
    TIMEOUT = 500
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
    for ttl := 1; ttl <= MAX_TTL; ttl++ {
        resIP, err := tracerouteWithGivenTTL(getIPv4Address(address, RECV_PORT), ttl)
        if err != nil && err.Error() != "Done" {
            if err.Error() == "operation not permitted" {
                fmt.Println("Operation not permitted, please run as admin")
                return
            }
            fmt.Printf("%d. *\r\n\r\n", ttl)
        } else {
            fmt.Printf("%d. %s\r\n", ttl, resIP)
            if isLocal(resIP) {
                fmt.Printf("local\r\n\r\n")
                continue
            }
            prevErr := err
            netName, origin, country, err := whois(resIP)
            if err == nil {
                fmt.Printf("%s %s %s\r\n\r\n", netName, origin, country)
            } else {
                fmt.Printf("\r\n")
            }
            if prevErr != nil {
                return
            }
        }
    }
}

func socketAddr() (addr [4]byte, err error) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return
    }
    for _, a := range addrs {
        if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if len(ipnet.IP.To4()) == net.IPv4len {
                copy(addr[:], ipnet.IP.To4())
                return
            }
        }
    }
    err = errors.New("You do not appear to be connected to the Internet")
    return
}


func tracerouteWithGivenTTL(address *syscall.SockaddrInet4, ttl int) (string, error) {
    writeSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

    if err != nil {
        return "", err
    }

    readSock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
    if err != nil {
        return "", err
    }

    tv := syscall.NsecToTimeval(1e6 * TIMEOUT)
    err = syscall.SetsockoptTimeval(readSock, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
    if err != nil {
		return "", err
	}

    err = syscall.SetsockoptInt(writeSock, 0, syscall.IP_TTL, ttl)
    if err != nil {
        return "", err
    }

    defer syscall.Close(readSock)
    defer syscall.Close(writeSock)

    sockAddr, err := socketAddr()
    if err != nil {
        return "", err
    }
    err = syscall.Bind(readSock, getIPv4AddressFromBytes(sockAddr, RECV_PORT))
    if err != nil {
        return "", err
    }
    err = syscall.Sendto(writeSock, []byte{0x0}, 0, address)
    if err != nil {
        return "", err
    }

    buf := make([]byte, 512)
	_, from, err := syscall.Recvfrom(readSock, buf, 0)
	if err != nil {
		return "", err
	}

    // tmp := from.(*syscall.SockaddrInet4).Addr

    // result := fmt.Sprintf("%d.%d.%d.%d", tmp[0], tmp[1], tmp[2], tmp[3])
    result := getIPString(from)
    if result ==  getIPString(address) {
        return result, errors.New("Done")
    }

    return result, nil
}

func getIPString(addr syscall.Sockaddr) string {
    tmp := addr.(*syscall.SockaddrInet4).Addr
    return fmt.Sprintf("%d.%d.%d.%d", tmp[0], tmp[1], tmp[2], tmp[3])

}

func getIPv4Address(hostname string, port int) *syscall.SockaddrInet4 {
    ipArray := net.ParseIP(hostname).To4()
    return getIPv4AddressFromBytes([4]byte{ipArray[0], ipArray[1], ipArray[2], ipArray[3]}, port)
}

func getIPv4AddressFromBytes(hostname[4] byte, port int) *syscall.SockaddrInet4 {
    return &syscall.SockaddrInet4 {
        Port: port,
        Addr: hostname,
    }
}
