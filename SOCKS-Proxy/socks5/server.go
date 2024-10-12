package socks5

import (
	"fmt"
	"log"
	"net"
	"socks-proxy/socks5/connection"
	"socks-proxy/socks5/handlers"

	"golang.org/x/sys/unix"
)

func StartServer(portNum int) error {
	listenFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf("Ошибка при создании сокета: %v", err)
	}

	err = unix.SetNonblock(listenFd, true)
	if err != nil {
		return fmt.Errorf("Ошибка при установке неблокирующего режима: %v", err)
	}

	addr := &unix.SockaddrInet4{Port: portNum}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())

	err = unix.Bind(listenFd, addr)
	if err != nil {
		return fmt.Errorf("Ошибка при привязке сокета: %v", err)
	}

	err = unix.Listen(listenFd, unix.SOMAXCONN)
	if err != nil {
		return fmt.Errorf("Ошибка при прослушивании порта: %v", err)
	}

	// Создаем UDP-сокет для DNS
	dnsFd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		return fmt.Errorf("Ошибка при создании UDP-сокета: %v", err)
	}
	err = unix.SetNonblock(dnsFd, true)
	if err != nil {
		return fmt.Errorf("Ошибка при установке неблокирующего режима для UDP-сокета: %v", err)
	}

	// Адрес DNS-сервера (Google DNS)
	dnsServer := &unix.SockaddrInet4{Port: 53}
	copy(dnsServer.Addr[:], net.ParseIP("8.8.8.8").To4())

	log.Println("SOCKS5 прокси-сервер слушает на порту", portNum)

	// Мапа для хранения соединений
	connections := make(map[int]*connection.Connection)
	dnsQueries := make(map[uint16]*connection.Connection)

	// Добавляем слушающий сокет и DNS-сокет в список отслеживаемых
	pollFds := []unix.PollFd{
		{Fd: int32(listenFd), Events: unix.POLLIN},
		{Fd: int32(dnsFd), Events: unix.POLLIN},
	}

	for {
		_, err := unix.Poll(pollFds, -1)
		if err != nil {
			return fmt.Errorf("Ошибка при вызове Poll: %v", err)
		}

		// Используем копию pollFds, чтобы избежать проблем с изменением слайса во время итерации
		pollFdsCopy := make([]unix.PollFd, len(pollFds))
		copy(pollFdsCopy, pollFds)

		for i := 0; i < len(pollFdsCopy); i++ {
			pfd := &pollFdsCopy[i]

			if pfd.Revents == 0 {
				continue
			}

			if pfd.Fd == int32(listenFd) {
				// Новое входящее соединение
				handlers.HandleNewConnection(listenFd, &pollFds, connections)
			} else if pfd.Fd == int32(dnsFd) {
				// Обработка DNS-ответа
				handlers.HandleDNSResponse(dnsFd, &pollFds, connections, dnsQueries)
			} else {
				// Обработка данных на клиентских и серверных соединениях
				conn := connections[int(pfd.Fd)]
				if conn == nil {
					continue
				}

				if pfd.Revents&(unix.POLLHUP|unix.POLLERR|unix.POLLNVAL) != 0 {
					// Соединение закрыто или ошибка
					handlers.CloseConnection(conn, &pollFds, connections)
					continue
				}

				handlers.HandleConnectionEvent(pfd, conn, &pollFds, connections, dnsFd, dnsServer, dnsQueries)
			}
		}
	}
}
