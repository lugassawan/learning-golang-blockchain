package blockchain

import (
	"fmt"
	"log"
	"os"

	"github.com/lugassawan/learning-golang-blockchain/transaction"
	"github.com/lugassawan/learning-golang-blockchain/utils"
	"go.etcd.io/bbolt"
)

const (
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type Blockchain struct {
	tip []byte
	db  *bbolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string, nodeId string) *Blockchain {
	if utils.CheckDB(nodeId) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bbolt.Open(utils.GetDBPath(nodeId), 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash(), genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash())
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash()
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(nodeId string) *Blockchain {
	if !utils.CheckDB(nodeId) {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bbolt.Open(utils.GetDBPath(nodeId), 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash())

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash(), blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height() > lastBlock.Height() {
			err = b.Put([]byte("l"), block.Hash())
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash()
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

func (bc *Blockchain) Close() {
	bc.db.Close()
}
