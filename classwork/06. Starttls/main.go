package main

import "net"
import "fmt"
import "bufio"
import "log"
import "regexp"
import "strings"
import "crypto/tls"
import "encoding/base64"

const (
    ADDR = "smtp.mail.ru"
    PORT = 587
    LOGIN = "test321"
    PASSWORD = "test123"
)

func main() {
    msg := ""
    conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", ADDR, PORT))
    msg, _ = read(conn)
    
    write(conn, "ehlo localhost")
    msg, _ = read(conn)
    
    write(conn, "starttls")
    msg, _ = read(conn)

    TLSconfig := &tls.Config{ InsecureSkipVerify: true }
    sconn := tls.Client(conn, TLSconfig)
    
    write(sconn, "ehlo localhost")
    msg, _ = read(sconn)

    write(sconn, "auth login")
    msg, _ = read(sconn)
    
    write(sconn, base64.StdEncoding.EncodeToString([]byte(LOGIN)))
    msg, _ = read(sconn)
    write(sconn, base64.StdEncoding.EncodeToString([]byte(PASSWORD)))
    msg, _ = read(sconn)
    
    write(sconn, "quit")
    msg, _ = read(sconn)
    log.Println(msg)
}

func write(conn net.Conn, message string) error {
    fmt.Fprintf(conn, message + "\n")
    log.Println("[->]    ", message)
    return nil
}

func read(conn net.Conn) (string, error) {
    result := ""
    buf := bufio.NewReader(conn)

    for {
        message, err := buf.ReadString('\n')
        log.Println("[<-]    ", strings.TrimSpace(message))
        if err != nil {
            return "", err
        }
        result += message
        if is_last_line(message) {
            break
        }
    }
    log.Println()
    return result, nil
}

func is_last_line(line string) bool {
    last_line_format := regexp.MustCompile(`^\d{3} .*`)
    if len(line) < 2 {
        return false
    }
    return last_line_format.MatchString(line)
}
