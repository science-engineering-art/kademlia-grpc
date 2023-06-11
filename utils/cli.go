package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetIpFromHost() string {
	cmd := exec.Command("hostname", "-i")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		//fmt.Println("Error running docker inspect:", err)
		return ""
	}
	ip := strings.TrimSpace(out.String())
	return ip
}

func DisplayHelp() {
	fmt.Println(`
help - This message
store <message> - Store a message on the network
get <key> - Get a message from the network
info - Display information about this node
	`)
}
