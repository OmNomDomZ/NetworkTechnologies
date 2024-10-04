package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			conn.Write([]byte("failure"))
		}
	}()

	fileNameBuf := make([]byte, serverFileNameMaxLength)
	_, err := conn.Read(fileNameBuf)
	if err != nil {
		fmt.Println("Error reading file name:", err.Error())
		return
	}
	fileName := strings.TrimRight(string(fileNameBuf), "\x00")

	var fileSize int64
	err = binary.Read(conn, binary.LittleEndian, &fileSize)
	if err != nil {
		fmt.Println("Error reading file size:", err.Error())
		return
	}
	fmt.Printf("Receiving file: %s, size: %d bytes\n", fileName, fileSize)

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
	startTime := time.Now()

	go getSpeed(conn, &receiveBytes, fileName, fileSize, startTime)

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

	fmt.Printf("File %s saved, received %d bytes\n", fileName, receiveBytes)

	checkSize(conn, receiveBytes, fileSize)
}

func checkSize(conn net.Conn, receiveBytes int64, fileSize int64) {
	if receiveBytes == fileSize {
		_, err := conn.Write([]byte("success"))
		if err != nil {
			fmt.Println("Error writing:", err.Error())
		}
	} else {
		_, err := conn.Write([]byte("failure"))
		if err != nil {
			fmt.Println("Error writing:", err.Error())
		}
	}
}

func getSpeed(conn net.Conn, receiveBytes *int64, fileName string, fileSize int64, startTime time.Time) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			speed := float64(*receiveBytes) / elapsed
			fmt.Printf("File: %s | Instant speed: %.2f bytes/sec | Received: %d bytes\n", fileName, speed, *receiveBytes)

			if *receiveBytes >= fileSize {
				checkSize(conn, *receiveBytes, fileSize)
				return
			}
		}
	}
}
