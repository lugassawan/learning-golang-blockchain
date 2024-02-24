package cli

import (
	"fmt"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/utils"
)

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if utils.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}

	cli.svc.Start(minerAddress)
}
