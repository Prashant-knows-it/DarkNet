package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// aesConn wraps a TCP connection in AES-256 stream encryption
type aesConn struct {
	net.Conn
	rStream cipher.Stream
	wStream cipher.Stream
}

// Read decrypts incoming bytes from the previous hop
func (c *aesConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	if n > 0 {
		c.rStream.XORKeyStream(p[:n], p[:n])
	}
	return
}

// Write encrypts outgoing bytes going back to the previous hop
func (c *aesConn) Write(p []byte) (n int, err error) {
	out := make([]byte, len(p))
	c.wStream.XORKeyStream(out, p)
	return c.Conn.Write(out)
}

// wrapNodeConn initializes the AES engine for a relay node
func wrapNodeConn(c net.Conn, key string) net.Conn {
	hash := sha256.Sum256([]byte(key))
	block, _ := aes.NewCipher(hash[:])
	iv := hash[:16] // In production, this must be randomly generated per session

	return &aesConn{
		Conn:    c,
		rStream: cipher.NewCFBDecrypter(block, iv), // Decrypt traffic coming FROM client
		wStream: cipher.NewCFBEncrypter(block, iv), // Encrypt traffic going TO client
	}
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run main.go <Name> <Key> <ListenPort> <NextHop>")
		return
	}

	name := os.Args[1]
	key := os.Args[2]
	listenPort := os.Args[3]
	nextHop := os.Args[4]

	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}

	fmt.Printf("[*] %s (AES-256) | Listening: %s | Next Hop: %s\n", name, listenPort, nextHop)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, nextHop, name, key)
	}
}

func handleConnection(clientConn net.Conn, nextHop, name, key string) {
	defer clientConn.Close()
	source := clientConn.RemoteAddr().String()

	nextConn, err := net.Dial("tcp", nextHop)
	if err != nil {
		fmt.Printf("[-] %s: Failed to reach next hop %s\n", name, nextHop)
		return
	}
	defer nextConn.Close()

	fmt.Printf("[+] %s: Intercepted connection from %s -> Peeling AES layer -> Forwarding to %s\n", name, source, nextHop)

	secureConn := wrapNodeConn(clientConn, key)

	go io.Copy(nextConn, secureConn)
	io.Copy(secureConn, nextConn)
}