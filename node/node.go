package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

const DefaultHTTPort = 8080

type (
	Node struct {
		state state
		port  uint64

		knownPeers map[string]PeerNode
	}
	state interface {
		AddTrx(db.Trx) error
		Persist() (db.Hash, error)
		LatestBlock() db.Block
		LatestBlockHash() db.Hash
		Balances() map[db.Account]uint64
	}
	PeerNode struct {
		IP         string `json:"ip"`
		Port       uint64 `json:"port"`
		IsBoostrap bool   `json:"is_bootstrap"`
		IsActive   bool   `json:"is_active"`
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
	TrxPostRes struct {
		Hash db.Hash `json:"block_hash"`
	}
	StatusRes struct {
		Hash       db.Hash             `json:"block_hash"`
		Height     uint64              `json:"block_height"`
		KnownPeers map[string]PeerNode `json:"peers_known"`
	}
)

func New(s state, port uint64, bootstrap PeerNode) *Node {
	knownPeers := map[string]PeerNode{
		bootstrap.Address(): bootstrap,
	}
	return &Node{s, port, knownPeers}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (n *Node) Run() error {
	mx := http.NewServeMux()

	mx.HandleFunc("/balances/list", n.GetBalances)
	mx.HandleFunc("/trx/add", n.PostTrx)
	mx.HandleFunc("/node/status", n.Status)

	fmt.Printf("Listening on %s:%d", "127.0.0.1\n", n.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), mx)
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

	if err := n.state.AddTrx(trx); err != nil {
		writeErr(w, err)
		return
	}

	hash, err := n.state.Persist()
	if err != nil {
		writeErr(w, err)
		return
	}

	res := TrxPostRes{hash}

	writeRes(w, res)
}

func (n *Node) Status(w http.ResponseWriter, r *http.Request) {
	res := StatusRes{
		Hash:       n.state.LatestBlockHash(),
		Height:     n.state.LatestBlock().Header.Height,
		KnownPeers: n.knownPeers,
	}
	writeRes(w, res)
}

func (pn PeerNode) Address() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
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
