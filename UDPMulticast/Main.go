package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	addr := os.Args[1]

	fmt.Printf("Get addr %v\n", addr)

	format := net.ParseIP(addr)

	if format == nil {
		fmt.Println("Invalid IP address")
		return
	}

	if format.To4() != nil {
		fmt.Println("IPv4")
	} else {
		fmt.Println("IPv6")
	}

}
