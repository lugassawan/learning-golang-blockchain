package blockchain

import (
	"fmt"

	"go.etcd.io/bbolt"
)

type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		fmt.Println("DB View Err : " + err.Error())
	}

	i.currentHash = block.PrevBlockHash()
	return block
}
