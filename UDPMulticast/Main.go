package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	messageInterval = 2 * time.Second // Интервал отправки сообщений
	timeoutDuration = 5 * time.Second // Время до признания копии мертвой
)

var (
	liveCopies = make(map[string]time.Time)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <program> <multicast address>")
		return
	}

	group := os.Args[1]
	port := "9999"

	// Присоединение к multicast-группе для приема сообщений
	addr, err := net.ResolveUDPAddr("udp", group+":"+port)
	if err != nil {
		panic(err)
	}

	// Используем multicast для получения сообщений
	listenConn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer listenConn.Close()

	// Отправка multicast-сообщения
	sendConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error setting up sender:", err)
		return
	}
	defer sendConn.Close()

	go func() {
		for {
			_, err = sendConn.Write([]byte("Hello Multicast from Go!"))
			if err != nil {
				fmt.Println("Error writing to UDP:", err)
				return
			}

			time.Sleep(messageInterval)
		}
	}()

	go receiveMessages(listenConn)

	// Периодическое обновление списка живых копий
	for {
		time.Sleep(timeoutDuration)
		printLiveCopies()
	}
}

// Функция для приема сообщений и обновления списка живых копий
func receiveMessages(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}
		message := strings.TrimSpace(string(buf[:n]))
		if strings.HasPrefix(message, "Hello Multicast") {
			liveCopies[src.String()] = time.Now()
		}
	}
}

// Функция для вывода списка живых копий
func printLiveCopies() {
	fmt.Println("Live copies:")
	for ip, lastSeen := range liveCopies {
		if time.Since(lastSeen) > timeoutDuration {
			delete(liveCopies, ip)
			continue
		}
		fmt.Printf("- %s\n", ip)
	}
	fmt.Println()
}
