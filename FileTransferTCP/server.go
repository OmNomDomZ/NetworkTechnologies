package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const serverFileNameMaxLength = 4096
const uploadDir = "./uploads"

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Server is waiting")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	fileNameBuf := make([]byte, serverFileNameMaxLength)
	_, err := conn.Read(fileNameBuf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	fileName := strings.TrimRight(string(fileNameBuf), "\x00")

	var fileSize int64
	err = binary.Read(conn, binary.LittleEndian, &fileSize)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating upload directory:", err.Error())
		return
	}

	filePath := filepath.Join(uploadDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err.Error())
		return
	}
	defer file.Close()

	buff := make([]byte, 1024)
	var receiveBytes int64
	for {
		n, err := conn.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading:", err.Error())
			return
		}
		receiveBytes += int64(n)
		file.Write(buff[:n])
	}

	fmt.Printf("File %s saved, get %d bytes\n", fileName, receiveBytes)
}
