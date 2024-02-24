package cli

import (
	"fmt"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/chainstate"
	"github.com/lugassawan/learning-golang-blockchain/utils"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !utils.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.Close()

	UTXOSet := chainstate.NewUTXOSet(bc)
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
