package main

import (
	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/cli"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.Close()

	cli := cli.NewCli(bc)
	cli.Run()
}
