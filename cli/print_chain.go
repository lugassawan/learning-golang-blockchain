package cli

import (
	"fmt"
	"strconv"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
)

func (cli *CLI) printChain(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Close()

	iterator := bc.Iterator()

	for {
		block := iterator.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash())
		fmt.Printf("Height: %d\n", block.Height())
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash())
		pow := blockchain.NewProofOfWork(block)

		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions() {
			fmt.Println(tx)
		}

		fmt.Printf("\n\n")

		if len(block.PrevBlockHash()) == 0 {
			break
		}
	}
}
