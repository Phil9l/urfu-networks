package main

import (
    "strings"
    "fmt"
)

const ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func main() {
	
}

func base64encode(data string) string {
    extraLen := (3 - (len(data) % 3)) % 3
    data += strings.Repeat("\000", extraLen)
    result := ""
    
    for i := 0; i < len(data); i += 3 {
        buf := (int(data[i]) << 16) + (int(data[i + 1]) << 8) + int(data[i + 2])
        for j := 18; j >= 0; j -= 6 {
            result += fmt.Sprintf("%c", ALPHABET[(buf >> uint(j)) & 63])
        }
    }
    
    return result[:len(result)-extraLen] + strings.Repeat("=", extraLen)
}

