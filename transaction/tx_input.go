package transaction

import (
	"bytes"

	"github.com/lugassawan/learning-golang-blockchain/utils"
)

// TXInput represents a transaction input
type TXInput struct {
	txId      []byte
	vout      int
	signature []byte
	pubkey    []byte
}

// NewTXInput create a new TXInput
func NewTXInput(txId []byte, vout int, signature []byte, publicKey []byte) *TXInput {
	return &TXInput{txId, vout, signature, publicKey}
}

func (ti *TXInput) TxId() []byte {
	return ti.txId
}

func (ti *TXInput) Vout() int {
	return ti.vout
}

func (ti *TXInput) Signature() []byte {
	return ti.signature
}

func (ti *TXInput) PubKey() []byte {
	return ti.pubkey
}

// UsesKey checks whether the address initiated the transaction
func (ti *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := utils.HashPubKey(ti.pubkey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
