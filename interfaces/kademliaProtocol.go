package interfaces

import "github.com/science-engineering-art/spotify/src/kademlia/structs"

type KademliaProtocol interface {
	Ping(structs.Node) structs.Node
	Store(value *[]byte) error
	FindNode(target []byte) (kBucket []structs.Node)
	FindValue(infoHash []byte) (value *[]byte, neighbors *[]structs.Node)
}
