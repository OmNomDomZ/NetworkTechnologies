package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const clientFileNameMaxLength = 4096

type FileHandler struct {
	conn      net.Conn
	fileName  string
	fileSize  int64
	startTime time.Time
	mutex     sync.Mutex
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: client <server> <port> <pathToFile>")
		return
	}

	server := os.Args[1]
	port := os.Args[2]
	filePath := os.Args[3]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	fileSize := fileInfo.Size()
	fileName := filepath.Base(filePath)

	conn, err := net.Dial("tcp", server+":"+port)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	paddedFileName := make([]byte, clientFileNameMaxLength)
	copy(paddedFileName, fileName)
	_, err = conn.Write(paddedFileName)
	if err != nil {
		fmt.Println("Error writing to server:", err)
		return
	}

	err = binary.Write(conn, binary.LittleEndian, fileSize)
	if err != nil {
		fmt.Println("Error writing to server:", err)
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from server:", err)
			return
		}
		conn.Write(buf[:n])
	}
}
