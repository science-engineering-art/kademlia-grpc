package structs

import (
	"bytes"
	"math/big"
)

// nodeList is used in order to sort a list of arbitrary nodes against a
// comparator. These nodes are sorted by xor distance
type ShortList struct {
	// Nodes are a list of nodes to be compared
	Nodes *[]Node

	// Comparator is the ID to compare to
	Comparator []byte
}

func areNodesEqual(n1 *Node, n2 *Node, allowNilID bool) bool {
	if n1 == nil || n2 == nil {
		return false
	}
	if !allowNilID {
		if n1.ID == nil || n2.ID == nil {
			return false
		}
		if !bytes.Equal(n1.ID, n2.ID) {
			return false
		}
	}
	if n1.IP != n2.IP {
		return false
	}
	if n1.Port != n2.Port {
		return false
	}
	return true
}

func (n *ShortList) RemoveNode(node *Node) {
	for i := 0; i < n.Len(); i++ {
		if bytes.Equal((*n.Nodes)[i].ID, node.ID) {
			*n.Nodes = append((*n.Nodes)[:i], (*n.Nodes)[i+1:]...)
			return
		}
	}
}

func (n *ShortList) Append(nodes []*Node) {
	//fmt.Println(nodes)
	for _, vv := range nodes {
		exists := false
		//fmt.Println(*n.Nodes)
		for _, v := range *n.Nodes {
			if bytes.Equal(v.ID, vv.ID) {
				exists = true
				break
			}
		}
		if !exists {
			*n.Nodes = append(*n.Nodes, *vv)
		}
	}
}

func (n *ShortList) Len() int {
	return len(*n.Nodes)
}

func (n *ShortList) Swap(i, j int) {
	(*n.Nodes)[i], (*n.Nodes)[j] = (*n.Nodes)[j], (*n.Nodes)[i]
}

func (n *ShortList) Less(i, j int) bool {
	iDist := getDistance((*n.Nodes)[i].ID, n.Comparator)
	jDist := getDistance((*n.Nodes)[j].ID, n.Comparator)

	return iDist.Cmp(jDist) == -1
}

func getDistance(id1 []byte, id2 []byte) *big.Int {
	buf1 := new(big.Int).SetBytes(id1)
	buf2 := new(big.Int).SetBytes(id2)
	result := new(big.Int).Xor(buf1, buf2)
	return result
}
