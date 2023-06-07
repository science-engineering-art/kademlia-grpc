package utils

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"strconv"

	"github.com/science-engineering-art/kademlia-grpc/pb"
	"github.com/science-engineering-art/kademlia-grpc/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// newID generates a new random ID
func NewID(ip string, port int) ([]byte, error) {
	hashValue := sha1.Sum([]byte(ip + ":" + strconv.FormatInt(int64(port), 10)))
	return []byte(hashValue[:]), nil
}

func GetFullNodeClient(ip *string, port *int) pb.FullNodeClient {
	address := fmt.Sprintf("%s:%d", *ip, *port)
	conn, _ := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	client := pb.NewFullNodeClient(conn)
	return client
}

func ClosestNodeToKey(key []byte, id1 []byte, id2 []byte) int {
	buf1 := new(big.Int).SetBytes(key)
	buf2 := new(big.Int).SetBytes(id1)
	buf3 := new(big.Int).SetBytes(id2)
	result1 := new(big.Int).Xor(buf1, buf2)
	result2 := new(big.Int).Xor(buf1, buf3)

	return result1.Cmp(result2)
}

func GetKBucketFromNodeArray(nodes *[]structs.Node) *pb.KBucket {
	result := pb.KBucket{Bucket: []*pb.Node{}}
	for _, node := range *nodes {
		result.Bucket = append(result.Bucket,
			&pb.Node{
				ID:   node.ID,
				IP:   node.IP,
				Port: int32(node.Port),
			},
		)
	}
	return &result
}
