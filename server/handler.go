package server

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/chainstate"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
)

func (s *server) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	s.knownNodes = append(s.knownNodes, payload.addrList...)
	fmt.Printf("There are %d known nodes now!\n", len(s.knownNodes))
	s.requestBlocks()
}

func (s *server) handleBlock(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.block
	block := blockchain.DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.sendGetData(payload.addrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := chainstate.NewUTXOSet(bc)
		UTXOSet.Reindex()
	}
}

func (s *server) handleInv(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.items), payload.kind)

	if payload.kind == "block" {
		s.blocksInTransit = payload.items

		blockHash := payload.items[0]
		s.sendGetData(payload.addrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range s.blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		s.blocksInTransit = newInTransit
	}

	if payload.kind == "tx" {
		txID := payload.items[0]

		if s.mempool[hex.EncodeToString(txID)].ID() == nil {
			s.sendGetData(payload.addrFrom, "tx", txID)
		}
	}
}

func (s *server) handleGetBlocks(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	s.sendInv(payload.addrFrom, "block", blocks)
}

func (s *server) handleGetData(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.kind == "block" {
		block, err := bc.GetBlock([]byte(payload.id))
		if err != nil {
			return
		}

		s.sendBlock(payload.addrFrom, &block)
	}

	if payload.kind == "tx" {
		txID := hex.EncodeToString(payload.id)
		tx := s.mempool[txID]

		s.sendTx(payload.addrFrom, tx)
		// delete(s.mempool, txID)
	}
}

func (s *server) handleTx(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.transaction
	tx := transaction.DeserializeTransaction(txData)
	s.mempool[hex.EncodeToString(tx.ID())] = &tx

	if s.nodeAddress == s.knownNodes[0] {
		for _, node := range s.knownNodes {
			if node != s.nodeAddress && node != payload.addFrom {
				s.sendInv(node, "tx", [][]byte{tx.ID()})
			}
		}
	} else {
		if len(s.mempool) >= 2 && len(s.miningAddress) > 0 {
		MineTransactions:
			var txs []*transaction.Transaction

			for id := range s.mempool {
				tx := s.mempool[id]
				if bc.VerifyTransaction(tx) {
					txs = append(txs, tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := transaction.NewCoinbaseTX(s.miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := chainstate.NewUTXOSet(bc)
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID())
				delete(s.mempool, txID)
			}

			for _, node := range s.knownNodes {
				if node != s.nodeAddress {
					s.sendInv(node, "block", [][]byte{newBlock.Hash()})
				}
			}

			if len(s.mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func (s *server) handleVersion(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.bestHeight

	if myBestHeight < foreignerBestHeight {
		s.sendGetBlocks(payload.addrFrom)
	} else if myBestHeight > foreignerBestHeight {
		s.sendVersion(payload.addrFrom, bc)
	}

	// sendAddr(payload.addrFrom)
	if !s.nodeIsKnown(payload.addrFrom) {
		s.knownNodes = append(s.knownNodes, payload.addrFrom)
	}
}

func (s *server) handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := s.bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		s.handleAddr(request)
	case "block":
		s.handleBlock(request, bc)
	case "inv":
		s.handleInv(request, bc)
	case "getblocks":
		s.handleGetBlocks(request, bc)
	case "getdata":
		s.handleGetData(request, bc)
	case "tx":
		s.handleTx(request, bc)
	case "version":
		s.handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}
