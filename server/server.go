package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:8888")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 1024)

	for {
		n, rAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Received %d bytes from %v\n", n, rAddr)
	}
}
