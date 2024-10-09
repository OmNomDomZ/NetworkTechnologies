package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: tcpserver <port>")
	}

	port := os.Args[1]

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("New connection accepted from " + conn.RemoteAddr().String())

	buf := make([]byte, 2)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	if buf[0] != 0x05 {
		log.Println("Unsupported SOCKS version:", buf[0])
		return
	}

	_, err = conn.Write([]byte{0x05, 0x00})
	if err != nil {
		log.Println("Error writing:", err)
		return
	}

	fmt.Println("SOCKS 5 handshake completed")
}
