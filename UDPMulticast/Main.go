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

	ip := net.ParseIP(group)
	if ip == nil {
		fmt.Println("Invalid multicast address")
		return
	}

	// Присоединение к multicast-группе для приема сообщений
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(group, port))
	if err != nil {
		panic(err)
	}

	// Используем multicast для получения сообщений
	listenConn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer listenConn.Close()

	// Поиск первого доступного интерфейса для IPv6 multicast
	var sendConn *net.UDPConn
	if ip.To4() == nil && addr.IP.IsMulticast() {
		iface, err := getFirstMulticastInterface()
		if err != nil {
			fmt.Println("Error finding multicast interface:", err)
			return
		}
		addr.Zone = iface.Name
	}

	// Соединение для отправки multicast-сообщения
	sendConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error setting up sender:", err)
		return
	}
	defer sendConn.Close()

	go sendMessages(sendConn)

	go receiveMessages(listenConn)

	// Периодическое обновление списка живых копий
	for {
		time.Sleep(timeoutDuration)
		printLiveCopies()
	}
}

func sendMessages(conn *net.UDPConn) {
	for {
		_, err := conn.Write([]byte("Hello"))
		if err != nil {
			fmt.Println("Error writing to UDP:", err)
			return
		}
		time.Sleep(messageInterval)
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
		if strings.HasPrefix(message, "Hello") {
			liveCopies[src.String()] = time.Now()
		}
	}
}

// Функция для вывода списка живых копий
func printLiveCopies() {
	fmt.Println("Live copies:")
	for ip, lastSeen := range liveCopies {
		if time.Since(lastSeen) > timeoutDuration {
			fmt.Printf("%v disconnected\n", ip)
			delete(liveCopies, ip)
			continue
		}
		fmt.Printf("- %s\n", ip)
	}
	fmt.Println()
}

// Получение первого доступного интерфейса для multicast
func getFirstMulticastInterface() (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagMulticast != 0 && iface.Flags&net.FlagUp != 0 {
			return &iface, nil
		}
	}
	return nil, fmt.Errorf("no multicast interface found")
}
