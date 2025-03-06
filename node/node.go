package node

import (
	"fmt"
	"net/http"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type (
	Node struct {
		state state
	}
	state interface {
		LatestBlockHash() db.Hash
		Balances() map[db.Account]uint64
	}
)

func New(s state) *Node {
	return &Node{s}
}

func (n *Node) Run() error {
	const port = 8080

	mx := http.NewServeMux()

	fmt.Printf("Listening on %s:%d", "127.0.0.1\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mx)
}
