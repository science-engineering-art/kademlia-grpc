package structs

import (
	"bytes"
	"strings"
)

type Node struct {
	ID   []byte `json:"id,omitempty"`
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (b Node) Equal(other Node) bool {
	return bytes.Equal(b.ID, other.ID) &&
		strings.EqualFold(b.IP, other.IP) &&
		b.Port == other.Port
}
