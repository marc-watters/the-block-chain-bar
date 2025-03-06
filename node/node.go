package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/marc-watters/the-block-chain-bar/v2/database"
	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type (
	Node struct {
		state state
	}
	state interface {
		AddTrx(db.Trx) error
		Persist() (db.Hash, error)
		LatestBlockHash() db.Hash
		Balances() map[db.Account]uint64
	}

	BalanceRes struct {
		Hash     db.Hash               `json:"block_hash"`
		Balances map[db.Account]uint64 `json:"balances"`
	}
	ErrRes struct {
		Error string `json:"error"`
	}
	TrxPostReq struct {
		From  db.Account `json:"from"`
		To    db.Account `json:"to"`
		Value uint64     `json:"value"`
		Data  string     `json:"data"`
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

func (n *Node) PostTrx(w http.ResponseWriter, r *http.Request) {
	var req TrxPostReq
	if err := readReq(r, &req); err != nil {
		writeErr(w, err)
		return
	}

	trx := db.NewTrx(req.From, req.To, req.Value, req.Data)

	if err := n.AddTrx(trx); err != nil {
		writeErr(w, err)
		return
	}

	hash, err := n.Persist()
	if err != nil {
		writeErr(w, err)
		return
	}

	res := struct {
		Hash database.Hash `json:"block_hash"`
	}{
		hash,
	}

	writeRes(w, res)
}

func writeRes(w http.ResponseWriter, data any) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		writeErr(w, err)
		return
	}

	if _, err := w.Write(dataJSON); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func writeErr(w http.ResponseWriter, err error) {
	errJSON, err := json.Marshal(ErrRes{err.Error()})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	http.Error(w, string(errJSON), http.StatusInternalServerError)
}

func readReq(r *http.Request, reqBody any) error {
	reqBodyJSON, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %v", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(reqBodyJSON, &reqBody); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %v", err)
	}

	return nil
}
