package cli

import (
	"fmt"
	"log"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/chainstate"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
	"github.com/lugassawan/learning-golang-blockchain/utils"
	"github.com/lugassawan/learning-golang-blockchain/wallet"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !utils.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !utils.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := chainstate.NewUTXOSet(bc)
	defer bc.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := wallet.CreateTransaction(to, amount, UTXOSet)

	if mineNow {
		cbTx := transaction.NewCoinbaseTX(from, "")
		txs := []*transaction.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		cli.svc.SendTx(cli.svc.KnownNodes()[0], tx)
	}

	fmt.Println("Success!")
}
