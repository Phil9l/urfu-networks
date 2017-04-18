package main

import (
    "bufio"
    "net"
    "fmt"
    "time"
    "strings"
    "errors"
)

const (
    IANA_ADDRESS = "whois.iana.org"
    WHOIS_PORT = 43
    CONNECT_TIMEOUT = 2000
    READ_TIMEOUT = 2000
    WRITE_TIMEOUT = 2000
)


func whois(hostname string) (string, string, string, error) {
    server, err := getWhoisServer(hostname)
    // log.Println(server)
    if err != nil {
        return "", "", "", err
    }

    connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", server, WHOIS_PORT), CONNECT_TIMEOUT * time.Millisecond)
    if err != nil {
        return "", "", "", err
    }
    connection.SetReadDeadline(time.Now().Add(READ_TIMEOUT * time.Millisecond))
    connection.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT * time.Millisecond))
    connection.Write([]byte(hostname + "\n"))
    bufReader := bufio.NewReader(connection)

    netName, origin, country := "", "", ""

    for true {
        data, err := bufReader.ReadBytes('\n')
        dataStr := string(data)

        if err != nil {
            break
        }
        // log.Println(strings.ToLower(dataStr))
        if strings.Contains(strings.ToLower(dataStr), "netname:") {
            netName = strings.Trim(dataStr[8:], " -\n")
        }
        if strings.Contains(strings.ToLower(dataStr), "origin:") {
            origin = strings.Trim(dataStr[7:], " -\n")
        }
        if strings.Contains(strings.ToLower(dataStr), "country:") {
            country = strings.Trim(dataStr[8:], " -\n")
        }
    }
    // log.Println(netName, origin, country)
    return netName, origin, country, nil
}

func isLocal(ip string) bool {
    return isIPInNetwork(ip, "10.0.0.0/8") || isIPInNetwork(ip, "172.16.0.0/12") || isIPInNetwork(ip, "192.168.0.0/16")
}

func isIPInNetwork(ip string, network string) bool {
    parsedIP, _, _ := net.ParseCIDR(ip + "/32")
    _, parsedNetwork, _ := net.ParseCIDR(network)

    return parsedNetwork.Contains(parsedIP)
}

func getWhoisServer(hostname string) (string, error) {
    connection, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", IANA_ADDRESS, WHOIS_PORT), CONNECT_TIMEOUT * time.Millisecond)
    if err != nil {
        return "", err
    }
    connection.SetReadDeadline(time.Now().Add(READ_TIMEOUT * time.Millisecond))
    connection.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT * time.Millisecond))
    connection.Write([]byte(hostname + "\n"))
    bufReader := bufio.NewReader(connection)

    for true {
        data, err := bufReader.ReadBytes('\n')
        dataStr := string(data)

        if err != nil {
            break
        }

        if strings.Contains(strings.ToLower(dataStr), "refer:") {
            return strings.Trim(dataStr[6:], " -\n"), nil
        }
    }

    return "", errors.New("No refer field.")
}
