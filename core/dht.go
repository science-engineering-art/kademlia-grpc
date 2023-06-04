package core

import (
	"bytes"

	"github.com/science-engineering-art/spotify/src/kademlia/interfaces"
	"github.com/science-engineering-art/spotify/src/kademlia/structs"
)

type DHT struct {
	structs.Node
	RoutingTable *structs.RoutingTable
	Storage      interfaces.Persistence
}

func (fn *DHT) Store(key []byte, data *[]byte) error {
	err := fn.Storage.Create(key, data)
	if err != nil {
		return err
	}
	return nil
}

func (fn *DHT) FindValue(infoHash *[]byte) (value *[]byte, neighbors *[]structs.Node) {
	value, err := fn.Storage.Read(*infoHash)
	if err != nil {
		//fmt.Println("Find Value error: ", err)
		neighbors = fn.RoutingTable.GetClosestContacts(structs.Alpha, *infoHash, []*structs.Node{&fn.Node}).Nodes
		return nil, neighbors
	}
	return value, nil
}

func (fn *DHT) FindNode(target *[]byte) (kBucket *[]structs.Node) {
	if bytes.Equal(fn.ID, *target) {
		kBucket = &[]structs.Node{fn.Node}
	}
	kBucket = fn.RoutingTable.GetClosestContacts(structs.Alpha, *target, []*structs.Node{&fn.Node}).Nodes

	return kBucket
}
