package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/science-engineering-art/kademlia-grpc/core"
	"github.com/science-engineering-art/kademlia-grpc/pb"
	"github.com/science-engineering-art/kademlia-grpc/structs"
	"github.com/science-engineering-art/kademlia-grpc/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gopkg.in/readline.v1"
)

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
		case "help":
			displayHelp()
		case "peer":
			if len(input) != 3 {
				displayHelp()
				continue
			}
			ip := input[1]
			port, _ := strconv.Atoi(input[2])

			// Create a gRPC server full node
			go CreateFullNodeServer(&ip, &port)

		case "store":
			if len(input) != 4 {
				displayHelp()
				continue
			}
			ip := input[1]
			port, _ := strconv.Atoi(input[2])
			data := input[3]

			client := GetFullNodeClient(&ip, &port)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			sender, err := client.Store(ctx)
			if err != nil {
				fmt.Println(err.Error())
			}
			keyHash := utils.GetSha1Hash(data)
			dataBytes := []byte(data)
			//fmt.Println("data bytes", dataBytes)
			err = sender.Send(&pb.StoreData{Key: keyHash, Value: &pb.Data{Init: 0, End: int64(len(dataBytes)), Buffer: dataBytes}})
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Stored ID: ", keyHash, "Stored Data:", string(dataBytes))

		case "ping":
			if len(input) != 5 {
				displayHelp()
				continue
			}
			ipSender := input[1]
			portSender, _ := strconv.Atoi(input[2])
			ipReceiver := input[3]
			portReceiver, _ := strconv.Atoi(input[4])
			client := GetFullNodeClient(&ipReceiver, &portReceiver)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			idSender, _ := utils.NewID(ipSender, portSender)
			pbNode, err := client.Ping(ctx, &pb.Node{ID: idSender, IP: ipSender, Port: int32(portSender)})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("The requested node is alive at:", pbNode.IP, ":", pbNode.Port)

		case "findnode":
			if len(input) != 4 {
				displayHelp()
				continue
			}
			ip := input[1]
			port, _ := strconv.Atoi(input[2])
			data := input[3]
			target := utils.GetSha1Hash(data)

			client := GetFullNodeClient(&ip, &port)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			pbKBucket, err := client.FindNode(ctx, &pb.Target{ID: target})
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("The found nodes where: ", pbKBucket.Bucket)

		case "findvalue":
			if len(input) != 4 {
				displayHelp()
				continue
			}
			ip := input[1]
			port, _ := strconv.Atoi(input[2])
			data := input[3]
			target := utils.GetSha1Hash(data)

			if len(target) == 0 {
				fmt.Println("Invalid target decoding.")
				continue
			}

			client := GetFullNodeClient(&ip, &port)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			receiver, err := client.FindValue(ctx, &pb.Target{ID: target})
			if err != nil {
				fmt.Println(err.Error())
			}

			buffer := []byte{}
			nearestNeighbors := []*pb.Node{}
			var init int64 = 0

			for {
				data, err := receiver.Recv()
				if data == nil {
					break
				}
				if len(data.KNeartestBuckets.Bucket) != 0 {
					nearestNeighbors = data.KNeartestBuckets.Bucket
					break
				}
				if init == data.Value.Init {
					buffer = append(buffer, data.Value.Buffer...)
					init = data.Value.End
				} else {
					fmt.Println(err.Error())
				}
			}
			foundValue := buffer
			if len(foundValue) == 0 {
				fmt.Println("Not found the requested value, this are alpha closest nodes:", nearestNeighbors)
			} else {
				fmt.Println("Found value:", foundValue)
			}

		case "lookup":
			if len(input) != 6 {
				displayHelp()
				continue
			}
			ip := input[1]
			port, _ := strconv.Atoi(input[2])
			bootIp := input[3]
			bootPort, _ := strconv.Atoi(input[4])
			data := input[5]
			target := utils.GetSha1Hash(data)

			grpcServerAddress := ip + ":" + strconv.FormatInt(int64(port), 10)
			fullNodeServer := *core.NewFullNode(ip, port, 0, structs.NewStorage(), false)

			// Create gRPC Server for ip and port
			go func() {
				grpcServer := grpc.NewServer()

				pb.RegisterFullNodeServer(grpcServer, &fullNodeServer)
				reflection.Register(grpcServer)

				listener, err := net.Listen("tcp", grpcServerAddress)
				if err != nil {
					log.Fatal("cannot create grpc server: ", err)
				}

				log.Printf("start gRPC server on %s", listener.Addr().String())
				err = grpcServer.Serve(listener)
				if err != nil {
					log.Fatal("cannot create grpc server: ", err)
				}
			}()

			//Send ping rpc from bootIp:bootPort for adding it to routing table as entry points
			client := GetFullNodeClient(&ip, &port)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			idSender, _ := utils.NewID(bootIp, bootPort)
			pbNode, err := client.Ping(ctx, &pb.Node{ID: idSender, IP: bootIp, Port: int32(bootPort)})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Made Ping from ", bootIp, ":", bootPort, "to", pbNode.IP, ":", pbNode.Port)
			nearestNodes, _ := fullNodeServer.LookUp(target)
			fmt.Println("This are the", structs.K, "closes node to", ip, ":", port, " ==> ", nearestNodes)
		}
	}
}

func displayHelp() {
	fmt.Println(`
help - This message
store <message> - Store a message on the network
get <key> - Get a message from the network
info - Display information about this node
	`)
}

func CreateFullNodeServer(ip *string, port *int) {
	grpcServerAddress := *ip + ":" + strconv.FormatInt(int64(*port), 10)
	fullNodeServer := *core.NewFullNode(*ip, *port, 0, structs.NewStorage(), false)

	// Create gRPC Server
	grpcServer := grpc.NewServer()

	pb.RegisterFullNodeServer(grpcServer, &fullNodeServer)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", grpcServerAddress)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}

	id, _ := utils.NewID(*ip, *port)
	log.Printf("start gRPC server on %s with id %s", listener.Addr().String(), base64.RawURLEncoding.EncodeToString(id))
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}
}

func GetFullNodeClient(ip *string, port *int) pb.FullNodeClient {
	address := fmt.Sprintf("%s:%d", *ip, *port)
	conn, _ := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	client := pb.NewFullNodeClient(conn)
	return client
}
