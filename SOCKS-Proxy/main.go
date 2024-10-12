package main

import (
	"fmt"
	"log"
	"os"
	"socks-proxy/socks5"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("socks5proxy <порт>")
		return
	}

	port := os.Args[1]
	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Некорректный номер порта:", err)
	}

	err = socks5.StartServer(portNum)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
