package main

import (
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func main() {
	// Configure TLS (skip certificate verification, only for development)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, // Don't use this option in production!
	}

	// Connect to server
	conn, err := tls.Dial("tcp", "localhost:8443", tlsConfig)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	// Create a buffered reader
	buf := make([]byte, 1024)
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')

		// Send message
		message := []byte(text + "\n")
		if _, err := conn.Write(message); err != nil {
			log.Fatalf("Write failed: %v", err)
		}

		// Read response
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Read failed: %v", err)
		}
		// transform the response to string as md5 hash representation

		hexHash := hex.EncodeToString(buf[:n])
		log.Printf("Server response MD5 hash: %s", hexHash)
	}
}
