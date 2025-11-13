package main

import (
	"crypto/md5"
	"crypto/tls"
	"log"
	"net"
)

func main() {
	// Load certificate and private key
	cert, err := tls.LoadX509KeyPair("localhost+2.pem", "localhost+2-key.pem")
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create TLS listener
	listener, err := tls.Listen("tcp", ":8443", tlsConfig)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("TLS server started, listening on :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read: %v", err)
			return
		}

		log.Printf("Received message: %s", string(buf[:n]))

		// calculate the md5 hash of the message and write it back to the client
		hash := md5.Sum(buf[:n])
		if _, err := conn.Write(hash[:]); err != nil {
			log.Printf("write: %v", err)
			return
		}
		log.Printf("Sent MD5 hash: %x", hash[:])
	}
	log.Printf("Closing connection: %s", conn.RemoteAddr())
}
