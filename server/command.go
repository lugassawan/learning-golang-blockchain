package server

import "fmt"

func (s *Server) commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func (s *Server) bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func (s *Server) extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func (s *Server) requestBlocks() {
	for _, node := range s.knownNodes {
		s.sendGetBlocks(node)
	}
}
