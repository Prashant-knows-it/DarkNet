package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type aesConn struct {
	net.Conn
	rStream cipher.Stream
	wStream cipher.Stream
}

func (c *aesConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	if n > 0 {
		c.rStream.XORKeyStream(p[:n], p[:n])
	}
	return
}

func (c *aesConn) Write(p []byte) (n int, err error) {
	out := make([]byte, len(p))
	c.wStream.XORKeyStream(out, p)
	return c.Conn.Write(out)
}

// wrapClientConn aligns the ciphers for the client side of the tunnel
func wrapClientConn(c net.Conn, key string) net.Conn {
	hash := sha256.Sum256([]byte(key))
	block, _ := aes.NewCipher(hash[:])
	iv := hash[:16]

	return &aesConn{
		Conn:    c,
		rStream: cipher.NewCFBDecrypter(block, iv), // Decrypt replies coming from nodes
		wStream: cipher.NewCFBEncrypter(block, iv), // Encrypt data going to nodes
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:10800")
	if err != nil {
		log.Fatalf("Failed to bind port: %v", err)
	}
	fmt.Println("[*] Shadow Gateway (AES-256) listening on 127.0.0.1:10800")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 256)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil { return }
	if buf[0] != 0x05 { return }
	io.ReadFull(conn, buf[:buf[1]])
	conn.Write([]byte{0x05, 0x00})
	if _, err := io.ReadFull(conn, buf[:4]); err != nil { return }
	if buf[1] != 0x01 { return }

	var target string
	if buf[3] == 0x03 {
		io.ReadFull(conn, buf[:1])
		domainLen := int(buf[0])
		domainBuf := make([]byte, domainLen)
		io.ReadFull(conn, domainBuf)
		portBuf := make([]byte, 2)
		io.ReadFull(conn, portBuf)
		port := binary.BigEndian.Uint16(portBuf)

		target = fmt.Sprintf("%s:%d", string(domainBuf), port)
		fmt.Printf("[+] SOCKS5 Request: %s\n", target)

		conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

		if string(domainBuf) == "mysite.shadow" {
			node1Conn, err := net.Dial("tcp", "fcufk-27-61-117-202.run.pinggy-free.link:36063")
			if err != nil {
				fmt.Println("[-] Node 1 offline")
				return
			}
			defer node1Conn.Close()

			fmt.Println("[+] Wrapping payload in 3 AES-256 layers (K3 -> K2 -> K1)...")

			// ONION ROUTING CRYPTO STACK
			layer1 := wrapClientConn(node1Conn, "KEY_1")
			layer2 := wrapClientConn(layer1, "KEY_2")
			layer3 := wrapClientConn(layer2, "KEY_3")

			go io.Copy(layer3, conn) 
			io.Copy(conn, layer3)    
		}
	}
}