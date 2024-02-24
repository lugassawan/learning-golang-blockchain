package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
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

// BuildTransaction creates a coinbase transaction
func BuildTransaction(inputs []TXInput, outputs []TXOutput) *Transaction {
	tx := Transaction{nil, inputs, outputs}
	tx.id = tx.Hash()

	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
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

// Sign signs each input of a Transaction
func (t *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if t.IsCoinbase() {
		return
	}

	for _, vin := range t.Vin() {
		if prevTxs[hex.EncodeToString(vin.TxId())].id == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := t.TrimmedCopy()

	for inId, vin := range txCopy.Vin() {
		prevTx := prevTxs[hex.EncodeToString(vin.TxId())]
		txCopy.Vin()[inId].signature = nil
		txCopy.Vin()[inId].pubkey = prevTx.Vout()[vin.Vout()].PubKeyHash()

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		t.Vin()[inId].signature = signature
		txCopy.Vin()[inId].pubkey = nil
	}
}

// String returns a human-readable representation of a transaction
func (t *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", t.ID()))

	for i, input := range t.Vin() {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TxId()))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout()))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature()))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey()))
	}

	for i, output := range t.Vout() {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value()))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash()))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (t *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range t.vin {
		inputs = append(inputs, TXInput{vin.TxId(), vin.Vout(), nil, nil})
	}

	for _, vout := range t.vout {
		outputs = append(outputs, TXOutput{vout.Value(), vout.PubKeyHash()})
	}

	return Transaction{t.ID(), inputs, outputs}
}

// Verify verifies signatures of Transaction inputs
func (t *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if t.IsCoinbase() {
		return true
	}

	for _, vin := range t.vin {
		if prevTxs[hex.EncodeToString(vin.txId)].id == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := t.TrimmedCopy()
	curve := elliptic.P256()

	for inId, vin := range t.vin {
		prevTx := prevTxs[hex.EncodeToString(vin.TxId())]
		txCopy.Vin()[inId].signature = nil
		txCopy.Vin()[inId].pubkey = prevTx.Vout()[vin.Vout()].PubKeyHash()

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature())
		r.SetBytes(vin.Signature()[:(sigLen / 2)])
		s.SetBytes(vin.Signature()[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey())
		x.SetBytes(vin.PubKey()[:(keyLen / 2)])
		y.SetBytes(vin.PubKey()[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) {
			return false
		}

		txCopy.Vin()[inId].pubkey = nil
	}

	return true
}
