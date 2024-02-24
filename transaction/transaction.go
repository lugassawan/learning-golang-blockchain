package transaction

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

const subsidy = 10

type Transaction struct {
	id   []byte
	vin  []TXInput
	vout []TXOutput
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.id = tx.Hash()

	return &tx
}

func (t *Transaction) ID() []byte {
	return t.id
}

func (t *Transaction) Vin() []TXInput {
	return t.vin
}

func (t *Transaction) Vout() []TXOutput {
	return t.vout
}

// IsCoinbase checks whether the transaction is coinbase
func (t *Transaction) IsCoinbase() bool {
	return len(t.vin) == 1 && len(t.vin[0].txId) == 0 && t.vin[0].vout == -1
}

// Serialize returns a serialized Transaction
func (t *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(t)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (t *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *t
	txCopy.id = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
