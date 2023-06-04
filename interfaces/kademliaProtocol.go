package interfaces

import "github.com/science-engineering-art/kademlia-grpc/structs"

type KademliaProtocol interface {
	Ping(structs.Node) structs.Node
	Store(value *[]byte) error
	FindNode(target []byte) (kBucket []structs.Node)
	FindValue(infoHash []byte) (value *[]byte, neighbors *[]structs.Node)
}
