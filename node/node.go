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

	BalanceRes struct {
		Hash     db.Hash               `json:"block_hash"`
		Balances map[db.Account]uint64 `json:"balances"`
	}
)

func New(s state) *Node {
	return &Node{s}
}

func (n *Node) Run() error {
	const port = 8080

	mx := http.NewServeMux()

	mx.HandleFunc("/balances/list", n.GetBalances)

	fmt.Printf("Listening on %s:%d", "127.0.0.1\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mx)
}

func (n *Node) GetBalances(w http.ResponseWriter, r *http.Request) {
	res := BalanceRes{
		n.state.LatestBlockHash(),
		n.state.Balances(),
	}
	writeRes(w, res)
}

func writeRes(w http.ResponseWriter, data any) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(dataJSON); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
