package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/science-engineering-art/kademlia-grpc/pb"
	"github.com/science-engineering-art/kademlia-grpc/structs"
	"github.com/science-engineering-art/kademlia-grpc/utils"
	"google.golang.org/grpc"
)

type FullNodeClient struct {
	pb.FullNodeClient
	IP   string
	Port int
}

func NewClientNode(ip string, port int) *FullNodeClient {
	address := fmt.Sprintf("%s:%d", ip, port)

	grpcConn := make(chan grpc.ClientConn)

	fmt.Println("Before TLS Client")
	tlsCred, _ := utils.ClientTLSCredentials()

	go func() {
		fmt.Println("Try stablish communication")
		// stablish connection
		conn, _ := grpc.Dial(address,
			grpc.WithTransportCredentials(tlsCred))
		if conn == nil {
			fmt.Println("Connection failed!")
			return
		}
		fmt.Println("Stablish correct connection")
		grpcConn <- *conn
	}()

	select {
	case <-time.After(30 * time.Second):
		fmt.Println("n@@@@@@@@@@@@@@@@@@@@@@@@@")
		return nil
	case conn := <-grpcConn:
		client := pb.NewFullNodeClient(&conn)
		if client != nil {
			fmt.Println("OKKKKKKKKKK")
		} else {
			fmt.Println("NOOOOOOOOOO")
		}
		fnClient := FullNodeClient{
			FullNodeClient: client,
			IP:             ip,
			Port:           port,
		}
		return &fnClient
	}
}

func (fn *FullNodeClient) Ping(sender structs.Node) (*structs.Node, error) {
	nodeChnn := make(chan *pb.Node)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		node, err := fn.FullNodeClient.Ping(ctx,
			&pb.Node{
				ID:   sender.ID,
				IP:   sender.IP,
				Port: int32(sender.Port),
			})
		if err != nil {
			fmt.Println(err)
		}
		nodeChnn <- node
	}()

	select {
	case <-time.After(5 * time.Second):
		log := fmt.Sprintf("node (%s:%d) doesn't respond", fn.IP, fn.Port)
		return nil, errors.New(log)
	case node := <-nodeChnn:
		return &structs.Node{
			ID:   node.ID,
			IP:   node.IP,
			Port: int(node.Port),
		}, nil
	}
}

// func (fn *FullNodeClient) Store() {

// }/
