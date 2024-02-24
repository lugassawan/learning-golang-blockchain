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

type Server struct {
	nodeId          string
	nodeAddress     string
	miningAddress   string
	knownNodes      []string
	blocksInTransit [][]byte
	mempool         map[string]*transaction.Transaction
}

// InitServer creates Server instance with empty miner address
func InitServer(nodeId string) *Server {
	return &Server{
		nodeId,
		fmt.Sprintf("localhost:%s", nodeId),
		"",
		[]string{"localhost:3000"},
		[][]byte{},
		make(map[string]*transaction.Transaction),
	}
}

func (s *Server) KnownNodes() []string {
	return s.knownNodes
}

// Start starts a node
func (s *Server) Start(minerAddress string) {
	s.miningAddress = minerAddress

	ln, err := net.Listen(protocol, s.nodeAddress)
	if err != nil {
		log.Panic(err)
	}

	defer ln.Close()

	bc := blockchain.NewBlockchain(s.nodeId)

	if s.nodeAddress != s.knownNodes[0] {
		s.sendVersion(s.knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}

		go s.handleConnection(conn, bc)
	}
}

func (s *Server) gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func (s *Server) nodeIsKnown(addr string) bool {
	for _, node := range s.knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
