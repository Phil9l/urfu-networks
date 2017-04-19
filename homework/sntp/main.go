package main

import (
    "encoding/binary"
    "flag"
    "fmt"
    "log"
    "net"
    "time"
)

const FROM_1900_TO_1970 = 2208988800

type UDPInfo struct {
    buf []byte
    addr *net.UDPAddr
}

func main() {
    flag.Usage = func() {
		fmt.Printf("Usage:\n./main [options] hostname\nOptions:\n")
		flag.PrintDefaults()
	}
    var portShort = flag.Int("p", -1, "Port to listen. Default is 123.")
    var portLong = flag.Int("port", 123, "Port to listen. Default is 123.")

    var delayShort = flag.Int("d", 0, "Delay added to real time. Default is 0.")
    var delayLong = flag.Int("delay", 0, "Delay added to real time. Default is 0.")
    flag.Parse()
    
    port := *portShort
    if *portShort == -1 {
        port = *portLong
    }
    delay := *delayShort
    if *delayShort == 0 {
        delay = *delayLong
    }
    
    log.Printf("Starting SNTP server on %d port. Delay is %d seconds\n", port, delay)
    
    serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    if err != nil {
        log.Println(err)
        return
    }

    ln, err := net.ListenUDP("udp", serverAddr)
    if err != nil {
        log.Println(err)
        return
    }
    
    in := make(chan UDPInfo, 16)

    go handleUDP(in, ln, delay)
    defer ln.Close()

    for {
        buf := make([]byte, 1024) 
        n, addr, err := ln.ReadFromUDP(buf)
        if err != nil {
            fmt.Println("Error: ",err)
        }
        in <- UDPInfo{buf[:n], addr}
    }
}

func handleUDP(ch chan UDPInfo, ln *net.UDPConn, delay int) {
    for frame := range ch {
        resp := generateAnswer(frame.buf, delay)
        ln.WriteToUDP(resp, frame.addr)
    }
}

func getFlag(data uint32, size uint, offset uint) uint32 {
    offset = (uint)(32) - offset - size
    mask := (((uint32)(1) << size) - 1) << offset;
    return (data & mask) >> offset;
}

func makeFlag(data uint32, size uint, offset uint) uint32 {
    offset = (uint)(32) - offset - size
    return data << offset;    
}

func getRow(data[] byte, rowOffset int) uint32 {
    offset := rowOffset * 4
    return binary.BigEndian.Uint32(data[offset:offset + 4])
}

func parseFirstLine(row uint32) uint32 {
    result := (uint32)(0)

    LI := getFlag(row, 2, 0)
    VN := getFlag(row, 3, 2)
    Mode := getFlag(row, 3, 5)
    Stratum := getFlag(row, 8, 8)
    Poll := getFlag(row, 8, 16)
    Precision := getFlag(row, 8, 24)
    
    log.Printf("LI:    \t%v", LI)
    log.Printf("VN:    \t%v", VN)
    log.Printf("Mode:\t%v", Mode)
    log.Printf("Stratum:\t%v", Stratum)
    log.Printf("Poll:\t%v", Poll)
    log.Printf("Precision:\t%v", Precision)
    
    result |= makeFlag(0, 2, 0)
    result |= makeFlag(VN, 3, 2)
    result |= makeFlag(4, 3, 5)
    result |= makeFlag(1, 8, 8)
    result |= makeFlag(Poll, 8, 16)
    result |= makeFlag(0, 8, 24)

    return result
}

func splitTime(time int64) (uint32, uint32) {
    utime := uint64(time)
    part1 := uint32(utime  >> 32)
    part2 := uint32(utime)
    
    return part1, part2
}

func uint32ToBytes(data uint32) []byte {
    result := make([]byte, 4)
    binary.BigEndian.PutUint32(result, data)
    return result
}

func generateAnswer(data[] byte, delay int) []byte {
    result := make([]byte, 0)

    // 0 line response
    row := getRow(data, 0)
    currentResult := parseFirstLine(row)
    result = append(result, uint32ToBytes(currentResult)...)

    currentTime := uint32(time.Now().Unix() + FROM_1900_TO_1970 + int64(delay))
    fracTime := uint32(time.Now().Nanosecond() + FROM_1900_TO_1970)

    // 1 line response is 0 (Delay)
    result = append(result, uint32ToBytes(uint32(0))...)
    
    // 2 line response is 0 (Dispersion)
    result = append(result, uint32ToBytes(uint32(0))...)

    // 3 line response is 0 (Must be IP of sync source)
    result = append(result, uint32ToBytes(uint32(0))...)

    // 4-5 line response is 0 (Timestamp, when last synced(never?).)
    result = append(result, uint32ToBytes(currentTime)...)
    result = append(result, uint32ToBytes(uint32(0))...)
    // result = append(result, uint32ToBytes(currentUTime2)...)
    
    // 6-7 line response is copied from clients Transmit Timestamp
    result = append(result, uint32ToBytes(getRow(data, 10))...)
    result = append(result, uint32ToBytes(getRow(data, 11))...)
    
    // 8-9 line response is current timestamp.
    result = append(result, uint32ToBytes(currentTime)...)
    result = append(result, uint32ToBytes(fracTime)...)
    
    // 10-11 line response is 0 (Timestamp, when last synced(never?).)
    result = append(result, uint32ToBytes(currentTime)...)
    result = append(result, uint32ToBytes(fracTime)...)

    return result;
}
