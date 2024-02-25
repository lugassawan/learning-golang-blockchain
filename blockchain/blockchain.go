package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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

// GetDB returns instance of bbolt.DB
func (bc *Blockchain) GetDB() *bbolt.DB {
	return bc.db
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

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(id []byte) (transaction.Transaction, error) {
	iterator := bc.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions() {
			if bytes.Compare(tx.ID(), id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash()) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]transaction.TXOutputs {
	utxo := make(map[string]transaction.TXOutputs)
	spentTXOs := make(map[string][]int)

	iterator := bc.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions() {
			txId := hex.EncodeToString(tx.ID())

		Outputs:
			for outIdx, out := range tx.Vout() {
				if spentTXOs[txId] != nil {
					for _, spentOutIdx := range spentTXOs[txId] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := utxo[txId]
				outs.Add(out)
				utxo[txId] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin() {
					inTxId := hex.EncodeToString(in.TxId())
					spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Vout())
				}
			}
		}

		if len(block.PrevBlockHash()) == 0 {
			break
		}
	}

	return utxo
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height()
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	iterator := bc.Iterator()

	for {
		block := iterator.Next()
		blocks = append(blocks, block.Hash())

		if len(block.PrevBlockHash()) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("Error: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height()
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash(), newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash())
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash()

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin() {
		prevTx, err := bc.FindTransaction(vin.TxId())
		if err != nil {
			log.Panic(err)
		}

		prevTxs[hex.EncodeToString(prevTx.ID())] = prevTx
	}

	tx.Sign(privateKey, prevTxs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin() {
		prevTx, err := bc.FindTransaction(vin.TxId())
		if err != nil {
			log.Panic(err)
		}

		prevTxs[hex.EncodeToString(prevTx.ID())] = prevTx
	}

	return tx.Verify(prevTxs)
}

// Close closes db connection
func (bc *Blockchain) Close() {
	bc.db.Close()
}
