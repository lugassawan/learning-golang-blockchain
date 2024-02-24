package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"github.com/lugassawan/learning-golang-blockchain/blockchain"
	"github.com/lugassawan/learning-golang-blockchain/transaction"
)

const (
	protocol      = "tcp"
	nodeVersion   = 1
	commandLength = 12
)

type server struct {
	nodeAddress     string
	miningAddress   string
	knownNodes      []string
	blocksInTransit [][]byte
	mempool         map[string]*transaction.Transaction
}

// StartServer starts a node
func StartServer(nodeId, minerAddress string) {
	svc := server{
		fmt.Sprintf("localhost:%s", nodeId),
		minerAddress,
		[]string{"localhost:3000"},
		[][]byte{},
		make(map[string]*transaction.Transaction),
	}

	ln, err := net.Listen(protocol, svc.nodeAddress)
	if err != nil {
		log.Panic(err)
	}

	defer ln.Close()

	bc := blockchain.NewBlockchain(nodeId)

	if svc.nodeAddress != svc.knownNodes[0] {
		svc.sendVersion(svc.knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}

		go svc.handleConnection(conn, bc)
	}
}

func (s *server) gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func (s *server) nodeIsKnown(addr string) bool {
	for _, node := range s.knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
