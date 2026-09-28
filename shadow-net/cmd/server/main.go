package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// The hidden server listens on port 8080
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}

	fmt.Println("[*] Hidden Web Server running on 127.0.0.1:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	// Read the incoming request
	buf := make([]byte, 4096)
	conn.Read(buf)
	fmt.Println("[+] Received request on hidden server")

	// Send the response back
	html := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		"Connection: close\r\n\r\n" +
		"<h1>Welcome to the TRUE Shadow Network!</h1>"
	conn.Write([]byte(html))
}