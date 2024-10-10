package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Использование: socks5proxy <порт>")
		return
	}

	port := os.Args[1]

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
	defer listener.Close()

	log.Println("SOCKS5 прокси-сервер слушает на порту", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Ошибка при принятии соединения:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("Новое соединение от", conn.RemoteAddr())

	if err := socks5Handshake(conn); err != nil {
		log.Println("Ошибка при рукопожатии:", err)
		return
	}

	if err := handleRequest(conn); err != nil {
		log.Println("Ошибка при обработке запроса:", err)
		return
	}
}

func socks5Handshake(conn net.Conn) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}

	if buf[0] != 0x05 {
		return fmt.Errorf("неподдерживаемая версия SOCKS: %d", buf[0])
	}

	nMethods := int(buf[1])

	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}

	methodSupported := false
	for _, method := range methods {
		if method == 0x00 {
			methodSupported = true
			break
		}
	}

	if !methodSupported {
		conn.Write([]byte{0x05, 0xFF})
		return fmt.Errorf("метод аутентификации не поддерживается")
	}

	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	return nil
}

func handleRequest(conn net.Conn) error {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	if header[0] != 0x05 {
		return fmt.Errorf("неподдерживаемая версия SOCKS в запросе: %d", header[0])
	}

	if header[1] != 0x01 {
		return fmt.Errorf("неподдерживаемая команда: %d", header[1])
	}

	if header[2] != 0x00 {
		return fmt.Errorf("поле RSV должно быть 0")
	}

	// address type
	atyp := header[3]

	destAddr, destPort, err := parseAddress(conn, atyp)
	if err != nil {
		return err
	}

	log.Printf("Подключение к %s:%d", destAddr, destPort)

	targetConn, err := net.Dial("tcp", net.JoinHostPort(destAddr, fmt.Sprintf("%d", destPort)))
	if err != nil {
		reply := buildReply(0x05, 0x01, atyp)
		conn.Write(reply)
		return err
	}
	defer targetConn.Close()

	reply := buildReply(0x05, 0x00, atyp)
	if _, err := conn.Write(reply); err != nil {
		return err
	}

	forwardTraffic(conn, targetConn)
	return nil
}

func parseAddress(conn net.Conn, atyp byte) (string, int, error) {
	var destAddr string

	switch atyp {
	case 0x01: // IPv4
		addr := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", 0, err
		}
		destAddr = net.IP(addr).String()

	case 0x03: // Доменное имя
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", 0, err
		}
		domainLen := int(lenBuf[0])
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", 0, err
		}
		destAddr = string(domain)

	case 0x04: // IPv6
		addr := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", 0, err
		}
		destAddr = net.IP(addr).String()

	default:
		return "", 0, fmt.Errorf("неизвестный тип адреса: %d", atyp)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", 0, err
	}
	destPort := int(binary.BigEndian.Uint16(portBuf))

	return destAddr, destPort, nil
}

// Функция для построения ответа клиенту
func buildReply(ver, rep, atyp byte) []byte {
	return []byte{
		ver,
		rep,
		0x00,
		atyp,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	}
}

func forwardTraffic(src net.Conn, dst net.Conn) {
	done := make(chan struct{}, 2)

	// от клиента к серверу
	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	// от сервера к клиенту
	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
}
