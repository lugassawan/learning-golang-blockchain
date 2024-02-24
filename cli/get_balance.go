package cli

import (
	"fmt"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/chainstate"
	"github.com/lugassawan/learning-golang-blockchain/utils"
)

func (cli *CLI) getBalance(address, nodeID string) {
	if !utils.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Close()

	UTXOSet := chainstate.NewUTXOSet(bc)

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value()
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
