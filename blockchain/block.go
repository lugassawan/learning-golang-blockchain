package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/lugassawan/learning-golang-blockchain/transaction"
	"github.com/lugassawan/learning-golang-blockchain/utils"
)

type Block struct {
	timestamp     int64
	transactions  []*transaction.Transaction
	prevBlockHash []byte
	hash          []byte
	nonce         int
	height        int
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{}, 0)
}

// NewBlock creates and returns Block
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.hash = hash[:]
	block.nonce = nonce

	return block
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))

	if err := decoder.Decode(&block); err != nil {
		log.Panic(err)
	}

	return &block
}

func (b *Block) Timestamp() int64 {
	return b.timestamp
}

func (b *Block) Transactions() []*transaction.Transaction {
	return b.transactions
}

func (b *Block) PrevBlockHash() []byte {
	return b.prevBlockHash
}

func (b *Block) Hash() []byte {
	return b.hash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Height() int {
	return b.height
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(b); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := utils.NewMerkleTree(transactions)

	return mTree.RootNode().Data()
}
