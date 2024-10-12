package dns

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"
)

func BuildDNSQuery(domain string) ([]byte, uint16) {
	rand.Seed(time.Now().UnixNano())
	id := uint16(rand.Intn(65535))
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], id)
	header[2] = 0x01                           // Recursion Desired
	binary.BigEndian.PutUint16(header[4:6], 1) // QDCOUNT=1

	question := []byte{}
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		question = append(question, byte(len(label)))
		question = append(question, []byte(label)...)
	}
	question = append(question, 0x00)       // Terminating byte
	question = append(question, 0x00, 0x01) // QTYPE=A
	question = append(question, 0x00, 0x01) // QCLASS=IN

	return append(header, question...), id
}

func ParseDNSResponse(data []byte) (string, error) {
	if len(data) < 12 {
		return "", errors.New("недостаточно данных в DNS-ответе")
	}
	answers := int(binary.BigEndian.Uint16(data[6:8]))
	if answers < 1 {
		return "", errors.New("нет ответов в DNS-ответе")
	}
	idx := 12
	// Пропускаем вопрос
	for {
		if idx >= len(data) {
			return "", errors.New("недостаточно данных при парсинге вопроса")
		}
		length := int(data[idx])
		if length == 0 {
			idx++
			break
		}
		idx += length + 1
	}
	idx += 4 // QTYPE и QCLASS

	// Читаем ответы
	for i := 0; i < answers; i++ {
		if idx+10 > len(data) {
			return "", errors.New("недостаточно данных при парсинге ответа")
		}
		// Пропускаем NAME
		if data[idx]&0xC0 == 0xC0 {
			idx += 2
		} else {
			for data[idx] != 0 {
				idx++
			}
			idx++
		}
		atype := binary.BigEndian.Uint16(data[idx : idx+2])
		idx += 8 // TYPE, CLASS, TTL
		rdlen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
		idx += 2
		if idx+rdlen > len(data) {
			return "", errors.New("недостаточно данных для RDATA")
		}
		if atype == 1 { // A record
			ip := net.IP(data[idx : idx+rdlen]).String()
			return ip, nil
		}
		idx += rdlen
	}
	return "", errors.New("не найден A-запись в DNS-ответе")
}
