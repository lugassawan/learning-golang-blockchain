package chainstate

import (
	"encoding/hex"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
	"go.etcd.io/bbolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	blockchain *blockchain.Blockchain
}

// NewUTXOSet creates instance UTXOSet
func NewUTXOSet(blockchain *blockchain.Blockchain) *UTXOSet {
	return &UTXOSet{blockchain}
}

func (utx *UTXOSet) Blockchain() *blockchain.Blockchain {
	return utx.blockchain
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (utx *UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := utx.blockchain.GetDB()

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := transaction.DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs() {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value()
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (utx *UTXOSet) FindUTXO(pubKeyHash []byte) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput
	db := utx.blockchain.GetDB()

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := transaction.DeserializeOutputs(v)

			for _, out := range outs.Outputs() {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (utx *UTXOSet) CountTransactions() int {
	db := utx.blockchain.GetDB()
	counter := 0

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex rebuilds the UTXO set
func (utx *UTXOSet) Reindex() {
	db := utx.blockchain.GetDB()
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bbolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	UTXO := utx.blockchain.FindUTXO()

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (utx *UTXOSet) Update(block *blockchain.Block) {
	db := utx.blockchain.GetDB()

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, trx := range block.Transactions() {
			if !trx.IsCoinbase() {
				for _, vin := range trx.Vin() {
					updatedOuts := transaction.TXOutputs{}
					outsBytes := b.Get(vin.TxId())
					outs := transaction.DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs() {
						if outIdx != vin.Vout() {
							updatedOuts.Add(out)
						}
					}

					if len(updatedOuts.Outputs()) == 0 {
						err := b.Delete(vin.TxId())
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.TxId(), updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := transaction.TXOutputs{}
			for _, out := range trx.Vout() {
				newOutputs.Add(out)
			}

			err := b.Put(trx.ID(), newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
