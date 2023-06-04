package core

import (
	"context"
	"fmt"
	"time"

	"github.com/science-engineering-art/spotify/src/kademlia/pb"
	"github.com/science-engineering-art/spotify/src/kademlia/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FullNodeClient struct {
	pb.FullNodeClient
}

func NewClientNode(ip string, port int) *FullNodeClient {

	address := fmt.Sprintf("%s:%d", ip, port)

	// stablish connection
	conn, _ := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	client := pb.NewFullNodeClient(conn)

	fnClient := FullNodeClient{FullNodeClient: client}

	return &fnClient
}

func (fn *FullNodeClient) Ping(sender structs.Node) (*structs.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node, err := fn.FullNodeClient.Ping(ctx,
		&pb.Node{
			ID:   sender.ID,
			IP:   sender.IP,
			Port: int32(sender.Port),
		})
	if err != nil {
		return nil, err
	}

	return &structs.Node{
		ID:   node.ID,
		IP:   node.IP,
		Port: int(node.Port),
	}, nil
}
