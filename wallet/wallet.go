package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/chainstate"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
	"github.com/lugassawan/learning-golang-blockchain/utils"
)

const version = byte(0x00)

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := utils.NewKeyPair()
	return &Wallet{private, public}
}

func (w *Wallet) GetPrivateKey() ecdsa.PrivateKey {
	return w.PrivateKey
}

func (w *Wallet) GetPublicKey() []byte {
	return w.PublicKey
}

// GetAddress returns wallet address
func (w *Wallet) GetAddress() []byte {
	pubKeyHash := utils.HashPubKey(w.GetPublicKey())

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := utils.Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := utils.Base58Encode(fullPayload)

	return address
}

// CreateTransaction a new transaction
func (w *Wallet) CreateTransaction(to string, amount int, UTXOSet *chainstate.UTXOSet) *transaction.Transaction {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	pubKeyHash := utils.HashPubKey(w.GetPublicKey())
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			inputs = append(inputs, *transaction.NewTXInput(txID, out, nil, w.GetPublicKey()))
		}
	}

	// Build a list of outputs
	from := fmt.Sprintf("%s", w.GetAddress())
	outputs = append(outputs, *transaction.NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *transaction.NewTXOutput(acc-amount, from)) // a change
	}

	tx := transaction.BuildTransaction(inputs, outputs)
	UTXOSet.Blockchain().SignTransaction(tx, w.GetPrivateKey())

	return tx
}
