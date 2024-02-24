package server

type addr struct {
	addrList []string
}

type block struct {
	addrFrom string
	block    []byte
}

type getblocks struct {
	addrFrom string
}

type getdata struct {
	addrFrom string
	kind     string
	id       []byte
}

type inv struct {
	addrFrom string
	kind     string
	items    [][]byte
}

type tx struct {
	addFrom     string
	transaction []byte
}

type verzion struct {
	version    int
	bestHeight int
	addrFrom   string
}
