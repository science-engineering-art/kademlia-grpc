package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/science-engineering-art/kademlia-grpc/core"
	"github.com/science-engineering-art/kademlia-grpc/structs"
	"github.com/science-engineering-art/kademlia-grpc/utils"
	"gopkg.in/readline.v1"
)

var fullNode core.FullNode
var grpcServerAddress string

func main() {
	// Init CLI for using Full Node Methods
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}
		input := strings.Split(line, " ")
		switch input[0] {
		case "node":
			if len(input) != 4 {
				utils.DisplayHelp()
				continue
			}
			port, _ := strconv.Atoi(input[1])
			bPort, _ := strconv.Atoi(input[2])
			isB, _ := strconv.ParseBool(input[3])

			flag.Parse()

			storage := structs.NewStorage()

			ip := utils.GetIpFromHost()
			grpcServerAddress = ip + ":" + strconv.FormatInt(int64(port), 10)
			fullNode = *core.NewFullNode(ip, port, bPort, storage, isB)
			go fullNode.CreateGRPCServer(grpcServerAddress)

			fmt.Println("Node running at:", ip, ":", port)

		case "store":
			if len(input) != 3 {
				utils.DisplayHelp()
				continue
			}
			key := input[1]
			data := []byte(input[2])
			id, err := fullNode.StoreValue(key, &data)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Stored with ID: ", id)
		case "get":
			if len(input) != 2 {
				utils.DisplayHelp()
				continue
			}
			key := input[1]
			value, err := fullNode.GetValue(key)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("The retrived value is:", string(value))
		case "dht":
			fullNode.PrintRoutingTable()
		}
	}
}
