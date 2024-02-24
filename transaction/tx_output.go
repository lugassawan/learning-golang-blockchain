package transaction

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/utils"
)

// TXOutput represents a transaction output
type TXOutput struct {
	value      int
	pubkeyHash []byte
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

func (to *TXOutput) Value() int {
	return to.value
}

func (to *TXOutput) PubKeyHash() []byte {
	return to.pubkeyHash
}

// Lock signs the output
func (to *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	to.pubkeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (to *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(to.pubkeyHash, pubKeyHash) == 0
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	outputs []TXOutput
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

func (out *TXOutputs) Outputs() []TXOutput {
	return out.outputs
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// Add adds TXOutput
func (outs TXOutputs) Add(txOutput TXOutput) {
	outs.outputs = append(outs.outputs, txOutput)
}
