package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
)

func (s *server) sendVersion(addr string, blockchain *blockchain.Blockchain) {
	bestHeight := blockchain.GetBestHeight()
	payload := s.gobEncode(verzion{nodeVersion, bestHeight, s.nodeAddress})

	request := append(s.commandToBytes("version"), payload...)

	s.sendData(addr, request)
}

func (s *server) sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range s.knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.knownNodes = updatedNodes

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func (s *server) sendAddr(address string) {
	nodes := addr{s.knownNodes}
	nodes.addrList = append(nodes.addrList, s.nodeAddress)
	payload := s.gobEncode(nodes)
	request := append(s.commandToBytes("addr"), payload...)

	s.sendData(address, request)
}

func (s *server) sendBlock(addr string, b *blockchain.Block) {
	data := block{s.nodeAddress, b.Serialize()}
	payload := s.gobEncode(data)
	request := append(s.commandToBytes("block"), payload...)

	s.sendData(addr, request)
}

func (s *server) sendInv(address, kind string, items [][]byte) {
	inventory := inv{s.nodeAddress, kind, items}
	payload := s.gobEncode(inventory)
	request := append(s.commandToBytes("inv"), payload...)

	s.sendData(address, request)
}

func (s *server) sendGetBlocks(address string) {
	payload := s.gobEncode(getblocks{s.nodeAddress})
	request := append(s.commandToBytes("getblocks"), payload...)

	s.sendData(address, request)
}

func (s *server) sendGetData(address, kind string, id []byte) {
	payload := s.gobEncode(getdata{s.nodeAddress, kind, id})
	request := append(s.commandToBytes("getdata"), payload...)

	s.sendData(address, request)
}

func (s *server) sendTx(addr string, tnx *transaction.Transaction) {
	data := tx{s.nodeAddress, tnx.Serialize()}
	payload := s.gobEncode(data)
	request := append(s.commandToBytes("tx"), payload...)

	s.sendData(addr, request)
}
