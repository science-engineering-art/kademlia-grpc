package structs

import "net"

type Message struct {
	IP     net.IP
	Buffer *[]byte
}
