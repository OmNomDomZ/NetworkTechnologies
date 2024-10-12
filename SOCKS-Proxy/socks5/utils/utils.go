package utils

import (
	"encoding/binary"
	"errors"
	"net"
	"socks-proxy/socks5/connection"

	"golang.org/x/sys/unix"
)

func RemovePollFd(pollFds *[]unix.PollFd, fd int) {
	for i, pfd := range *pollFds {
		if pfd.Fd == int32(fd) {
			*pollFds = append((*pollFds)[:i], (*pollFds)[i+1:]...)
			break
		}
	}
}

func UpdateEvents(conn *connection.Connection, pollFds *[]unix.PollFd, events int16) {
	for i := range *pollFds {
		if (*pollFds)[i].Fd == int32(conn.Fd) {
			(*pollFds)[i].Events = events
		}
		if conn.PeerFd > 0 && (*pollFds)[i].Fd == int32(conn.PeerFd) {
			(*pollFds)[i].Events = events
		}
	}
}

func BuildReply(rep byte, atyp byte, addr string, port int) []byte {
	buf := []byte{0x05, rep, 0x00, atyp}
	if atyp == 0x01 {
		ip := net.ParseIP(addr).To4()
		buf = append(buf, ip...)
	} else if atyp == 0x03 {
		buf = append(buf, byte(len(addr)))
		buf = append(buf, addr...)
	}
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, uint16(port))
	buf = append(buf, portBuf...)
	return buf
}

func ParseAddress(fd int, atyp byte) (string, int, error) {
	switch atyp {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		n, err := unix.Read(fd, addr)
		if err != nil || n != 4 {
			return "", 0, errors.New("ошибка чтения IPv4 адреса")
		}
		ip := net.IP(addr).String()
		port, err := readPort(fd)
		if err != nil {
			return "", 0, err
		}
		return ip, port, nil
	case 0x03: // Доменное имя
		lenBuf := make([]byte, 1)
		n, err := unix.Read(fd, lenBuf)
		if err != nil || n != 1 {
			return "", 0, errors.New("ошибка чтения длины доменного имени")
		}
		domainLen := int(lenBuf[0])
		domain := make([]byte, domainLen)
		n, err = unix.Read(fd, domain)
		if err != nil || n != domainLen {
			return "", 0, errors.New("ошибка чтения доменного имени")
		}
		port, err := readPort(fd)
		if err != nil {
			return "", 0, err
		}
		return string(domain), port, nil
	default:
		return "", 0, errors.New("неизвестный тип адреса")
	}
}

func readPort(fd int) (int, error) {
	portBuf := make([]byte, 2)
	n, err := unix.Read(fd, portBuf)
	if err != nil || n != 2 {
		return 0, errors.New("ошибка чтения порта")
	}
	port := int(binary.BigEndian.Uint16(portBuf))
	return port, nil
}
