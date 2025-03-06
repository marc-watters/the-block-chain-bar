package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

func (n *Node) GetBalances(w http.ResponseWriter, r *http.Request) {
	res := struct {
		Hash     db.Hash               `json:"block_hash"`
		Balances map[db.Account]uint64 `json:"balances"`
	}{
		n.state.LatestBlockHash(),
		n.state.Balances(),
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Node.GetBalances() write error: %v", err)
		return
	}
}
