package net

import (
	"fmt"
	"net"
	"time"

	"github.com/science-engineering-art/kademlia-grpc/utils"
)

func Send(addr net.Addr, buffer *[]byte) {
	conn, err := net.Dial("tcp", addr.String())
	for err != nil {
		conn, err = net.Dial("tcp", addr.String())
	}
	defer conn.Close()
	// fmt.Println("Stablish connection")

	conn.Write(*buffer)
}

func Recv(recv chan *net.TCPConn, port int) {
	address := fmt.Sprintf("%s:%d", utils.GetIpFromHost(), port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", address)

	lis, err := net.ListenTCP("tcp", tcpAddr)
	for err != nil {
		lis, err = net.ListenTCP("tcp", tcpAddr)
	}
	defer lis.Close()

	tcpConn, _ := lis.AcceptTCP()

	select {
	case conn, closed := <-recv:
		if !closed {
			recv <- conn
		}
	case <-time.After(100 * time.Millisecond):
		recv <- tcpConn
	}
}
