package node

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

const (
	DefaultIP      = "127.0.0.1"
	DefaultHTTPort = 8080

	endpointBalances              = "/balances/list"
	endpointPostTrx               = "/trx/add"
	endpointStatus                = "/node/status"
	endpointSync                  = "/node/sync"
	endpointSyncQueryKeyFromBlock = "fromBlock"
)

type (
	Node struct {
		state state
		ip    string
		port  uint64

		knownPeers map[string]PeerNode
	}
	state interface {
		AddBlock(db.Block) (db.Hash, error)
		AddBlocks([]db.Block) error
		AddTrx(db.Trx) error
		Persist() (db.Hash, error)
		LatestBlock() db.Block
		LatestBlockHash() db.Hash
		Balances() map[db.Account]uint64
		DataDir() string
	}
	PeerNode struct {
		IP         string `json:"ip"`
		Port       uint64 `json:"port"`
		IsBoostrap bool   `json:"is_bootstrap"`
		IsActive   bool   `json:"is_active"`
	}
)

func New(s state, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := map[string]PeerNode{
		bootstrap.Address(): bootstrap,
	}
	return &Node{s, ip, port, knownPeers}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (n *Node) Run() error {
	mx := http.NewServeMux()

	mx.HandleFunc(endpointBalances, n.GetBalances)
	mx.HandleFunc(endpointPostTrx, n.PostTrx)
	mx.HandleFunc(endpointStatus, n.Status)
	mx.HandleFunc(endpointSync, n.Sync)

	go func() {
		if err := n.sync(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "Node.Run() sync error:", err)
		}
	}()

	fmt.Printf("Listening on %s:%d\n", n.ip, n.port)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", n.ip, n.port), mx)
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

	block := db.NewBlock(
		n.state.LatestBlockHash(),
		n.state.LatestBlock().Header.Height+1,
		uint64(time.Now().Unix()),
		[]db.Trx{trx},
	)

	hash, err := n.state.AddBlock(block)
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

func (n *Node) Sync(w http.ResponseWriter, r *http.Request) {
	reqHash := r.URL.Query().Get(endpointSyncQueryKeyFromBlock)

	var hash db.Hash
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErr(w, err)
		return
	}

	blocks, err := db.GetBlocksAfter(hash, n.state.DataDir())
	if err != nil {
		writeErr(w, err)
		return
	}

	writeRes(w, SyncRes{Blocks: blocks})
}

func (pn PeerNode) Address() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}
