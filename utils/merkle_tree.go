package utils

import (
	"crypto/sha256"
)

// MerkleTree represent a Merkle tree
type MerkleTree struct {
	rootNode *MerkleNode
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

func (mt *MerkleTree) RootNode() *MerkleNode {
	return mt.rootNode
}

// MerkleNode represent a Merkle tree node
type MerkleNode struct {
	left  *MerkleNode
	right *MerkleNode
	data  []byte
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.data = hash[:]
	} else {
		prevHashes := append(left.data, right.data...)
		hash := sha256.Sum256(prevHashes)
		mNode.data = hash[:]
	}

	mNode.left = left
	mNode.right = right

	return &mNode
}

func (mNode *MerkleNode) Left() *MerkleNode {
	return mNode.left
}

func (mNode *MerkleNode) Right() *MerkleNode {
	return mNode.right
}

func (mNode *MerkleNode) Data() []byte {
	return mNode.data
}
