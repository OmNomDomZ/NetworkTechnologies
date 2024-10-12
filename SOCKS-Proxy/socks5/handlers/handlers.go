package handlers

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"socks-proxy/dns"
	"socks-proxy/socks5/connection"
	"socks-proxy/socks5/utils"

	"golang.org/x/sys/unix"
)

func HandleNewConnection(listenFd int, pollFds *[]unix.PollFd, connections map[int]*connection.Connection) {
	for {
		nfd, _, err := unix.Accept(listenFd)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				break
			}
			log.Println("Ошибка при принятии соединения:", err)
			break
		}

		err = unix.SetNonblock(nfd, true)
		if err != nil {
			log.Println("Ошибка при установке неблокирующего режима:", err)
			unix.Close(nfd)
			continue
		}

		conn := &connection.Connection{
			Fd:     nfd,
			State:  connection.StateHandshake,
			Buffer: make([]byte, 0),
		}
		connections[nfd] = conn

		*pollFds = append(*pollFds, unix.PollFd{Fd: int32(nfd), Events: unix.POLLIN})
		log.Println("Новое соединение")
	}
}

func HandleDNSResponse(dnsFd int, pollFds *[]unix.PollFd, connections map[int]*connection.Connection, dnsQueries map[uint16]*connection.Connection) {
	buf := make([]byte, 512)
	n, _, err := unix.Recvfrom(dnsFd, buf, 0)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return
		}
		log.Println("Ошибка при получении DNS-ответа:", err)
		return
	}

	if n > 0 {
		dnsMsg := buf[:n]
		id := binary.BigEndian.Uint16(dnsMsg[0:2])
		conn := dnsQueries[id]
		if conn != nil {
			ip, err := dns.ParseDNSResponse(dnsMsg)
			if err != nil {
				log.Println("Ошибка при парсинге DNS-ответа:", err)
				return
			}
			conn.DestAddr = ip
			conn.State = connection.StateConnecting
			// Устанавливаем соединение с сервером
			establishServerConnection(conn, pollFds, connections)
			delete(dnsQueries, id)
		}
	}
}

func HandleConnectionEvent(pfd *unix.PollFd, conn *connection.Connection, pollFds *[]unix.PollFd, connections map[int]*connection.Connection, dnsFd int, dnsServer *unix.SockaddrInet4, dnsQueries map[uint16]*connection.Connection) {
	switch conn.State {
	case connection.StateHandshake:
		if pfd.Revents&unix.POLLIN != 0 {
			err := handleHandshake(conn)
			if err != nil {
				log.Println("Ошибка рукопожатия:", err)
				CloseConnection(conn, pollFds, connections)
			}
		}
	case connection.StateRequest:
		if pfd.Revents&unix.POLLIN != 0 {
			err := handleRequest(conn, dnsFd, dnsServer, dnsQueries, pollFds, connections)
			if err != nil {
				log.Println("Ошибка при обработке запроса:", err)
				CloseConnection(conn, pollFds, connections)
			}
		}
	case connection.StateConnecting:
		if pfd.Revents&(unix.POLLOUT|unix.POLLIN) != 0 {
			err := finishConnection(conn)
			if err != nil {
				log.Println("Ошибка при подключении к серверу:", err)
				CloseConnection(conn, pollFds, connections)
			} else {
				conn.State = connection.StateForwarding
				// Отправляем клиенту ответ о успешном подключении
				reply := utils.BuildReply(0x00, conn.AddrType, conn.DestAddr, conn.DestPort)
				_, err = unix.Write(conn.Fd, reply)
				if err != nil {
					log.Println("Ошибка при отправке ответа клиенту:", err)
					CloseConnection(conn, pollFds, connections)
				} else {
					utils.UpdateEvents(conn, pollFds, unix.POLLIN)
				}
			}
		}
	case connection.StateForwarding:
		if pfd.Fd == int32(conn.Fd) && pfd.Revents&unix.POLLIN != 0 {
			// Клиент -> Сервер
			err := forwardData(conn.Fd, conn.PeerFd)
			if err != nil {
				CloseConnection(conn, pollFds, connections)
			}
		} else if pfd.Fd == int32(conn.PeerFd) && pfd.Revents&unix.POLLIN != 0 {
			// Сервер -> Клиент
			err := forwardData(conn.PeerFd, conn.Fd)
			if err != nil {
				CloseConnection(conn, pollFds, connections)
			}
		}
	case connection.StateClosing:
		CloseConnection(conn, pollFds, connections)
	}
}

func establishServerConnection(conn *connection.Connection, pollFds *[]unix.PollFd, connections map[int]*connection.Connection) {
	serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Println("Ошибка при создании сокета для сервера:", err)
		return
	}
	err = unix.SetNonblock(serverFd, true)
	if err != nil {
		log.Println("Ошибка при установке неблокирующего режима для сервера:", err)
		unix.Close(serverFd)
		return
	}
	conn.PeerFd = serverFd
	// Подключаемся к серверу
	addr := &unix.SockaddrInet4{Port: conn.DestPort}
	ipBytes := net.ParseIP(conn.DestAddr).To4()
	copy(addr.Addr[:], ipBytes)
	err = unix.Connect(serverFd, addr)
	if err != nil && err != unix.EINPROGRESS {
		log.Println("Ошибка при подключении к серверу:", err)
		unix.Close(serverFd)
		conn.State = connection.StateClosing
		return
	}
	// Добавляем серверный сокет в pollFds и connections
	*pollFds = append(*pollFds, unix.PollFd{Fd: int32(serverFd), Events: unix.POLLOUT})
	connections[serverFd] = conn
}

func handleHandshake(conn *connection.Connection) error {
	buf := make([]byte, 2)
	n, err := unix.Read(conn.Fd, buf)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		return err
	}
	if n == 0 {
		return io.EOF
	}

	if buf[0] != 0x05 {
		return errors.New("неподдерживаемая версия SOCKS")
	}

	methodsLen := int(buf[1])
	methods := make([]byte, methodsLen)
	n, err = unix.Read(conn.Fd, methods)
	if err != nil {
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
		// Отправляем ответ о том, что методы не поддерживаются
		unix.Write(conn.Fd, []byte{0x05, 0xFF})
		return errors.New("метод аутентификации не поддерживается")
	}

	// Отправляем ответ о выборе метода
	unix.Write(conn.Fd, []byte{0x05, 0x00})
	conn.State = connection.StateRequest
	return nil
}

func handleRequest(conn *connection.Connection, dnsFd int, dnsServer *unix.SockaddrInet4, dnsQueries map[uint16]*connection.Connection, pollFds *[]unix.PollFd, connections map[int]*connection.Connection) error {
	header := make([]byte, 4)
	n, err := unix.Read(conn.Fd, header)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		return err
	}
	if n == 0 {
		return io.EOF
	}

	if header[0] != 0x05 {
		return errors.New("неподдерживаемая версия SOCKS в запросе")
	}

	if header[1] != 0x01 {
		return errors.New("неподдерживаемая команда")
	}

	if header[2] != 0x00 {
		return errors.New("поле RSV должно быть 0")
	}

	atyp := header[3]
	conn.AddrType = atyp

	destAddr, destPort, err := utils.ParseAddress(conn.Fd, atyp)
	if err != nil {
		return err
	}
	conn.DestPort = destPort

	if atyp == 0x03 {
		// Доменное имя, отправляем DNS-запрос
		conn.State = connection.StateConnecting
		query, id := dns.BuildDNSQuery(destAddr)
		conn.DnsQueryID = id
		dnsQueries[id] = conn
		err = unix.Sendto(dnsFd, query, 0, dnsServer)
		if err != nil {
			return err
		}
	} else if atyp == 0x01 {
		// IPv4 адрес
		conn.DestAddr = destAddr
		// Устанавливаем соединение с сервером
		conn.State = connection.StateConnecting
		establishServerConnection(conn, pollFds, connections)
	} else {
		return errors.New("тип адреса не поддерживается")
	}
	return nil
}

func finishConnection(conn *connection.Connection) error {
	_, err := unix.Getpeername(conn.PeerFd)
	if err != nil {
		if err == unix.ENOTCONN {
			return nil
		}
		return err
	}
	return nil
}

func forwardData(srcFd, dstFd int) error {
	buf := make([]byte, 4096)
	n, err := unix.Read(srcFd, buf)
	if err != nil {
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return nil
		}
		return err
	}
	if n == 0 {
		return io.EOF
	}

	totalWritten := 0
	for totalWritten < n {
		m, err := unix.Write(dstFd, buf[totalWritten:n])
		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				continue
			}
			return err
		}
		totalWritten += m
	}
	return nil
}

func CloseConnection(conn *connection.Connection, pollFds *[]unix.PollFd, connections map[int]*connection.Connection) {
	unix.Close(conn.Fd)
	if conn.PeerFd > 0 {
		unix.Close(conn.PeerFd)
		delete(connections, conn.PeerFd)
	}
	delete(connections, conn.Fd)
	utils.RemovePollFd(pollFds, conn.Fd)
	if conn.PeerFd > 0 {
		utils.RemovePollFd(pollFds, conn.PeerFd)
	}
}
