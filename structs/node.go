package structs

import (
	"strings"
)

type Node struct {
	ID   []byte `json:"id,omitempty"`
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (b *Node) Equal(other Node) bool {
	return strings.EqualFold(b.IP, other.IP) &&
		b.Port == other.Port
}
