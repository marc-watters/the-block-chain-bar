package node

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	endpointAddPeer               = "/node/peer"
	endpointAddPeerQueryKeyIP     = "ip"
	endpointAddPeerQueryKeyPort   = "port"

	mininingIntervalSeconds = 10
)

type (
	Node struct {
		info PeerNode

		state           state
		knownPeers      map[string]PeerNode
		pendingTRXs     map[string]db.Trx
		archivedTRXs    map[string]db.Trx
		newSyncedBlocks chan db.Block
		newPendingTRXs  chan db.Trx
		isMining        bool
	}
	state interface {
		AddBlock(db.Block) (db.Hash, error)
		AddBlocks([]db.Block) error
		AddTrx(db.Trx) error
		Persist() (db.Hash, error)
		LatestBlock() db.Block
		LatestBlockHash() db.Hash
		NextBlockHeight() uint64
		Balances() map[db.Account]uint64
		DataDir() string
	}
	PeerNode struct {
		IP         string `json:"ip"`
		Port       uint64 `json:"port"`
		IsBoostrap bool   `json:"is_bootstrap"`

		connected bool
	}
)

func New(s state, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := map[string]PeerNode{
		bootstrap.Address(): bootstrap,
	}
	return &Node{s, ip, port, knownPeers}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}

func (n *Node) Run() error {
	mx := http.NewServeMux()

	mx.HandleFunc(endpointBalances, n.GetBalances)
	mx.HandleFunc(endpointPostTrx, n.PostTrx)
	mx.HandleFunc(endpointStatus, n.Status)
	mx.HandleFunc(endpointSync, n.Sync)
	mx.HandleFunc(endpointAddPeer, n.AddPeer)

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
		n.state.NextBlockHeight(),
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

func (n *Node) AddPeer(w http.ResponseWriter, r *http.Request) {
	peerIP := r.URL.Query().Get(endpointAddPeerQueryKeyIP)
	peerPortRaw := r.URL.Query().Get(endpointAddPeerQueryKeyPort)

	peerPort, err := strconv.ParseUint(peerPortRaw, 10, 32)
	if err != nil {
		writeRes(w, AddPeerRes{false, err.Error()})
		return
	}

	peer := NewPeerNode(peerIP, peerPort, false, true)

	n.addPeer(peer)

	fmt.Println("Peer", peer.Address(), "was added to known peers")

	writeRes(w, AddPeerRes{true, ""})
}

func (n *Node) addPeer(p PeerNode) {
	n.knownPeers[p.Address()] = p
}

func (n *Node) deletePeer(p PeerNode) {
	delete(n.knownPeers, p.Address())
}

func (n *Node) isKnownPeer(p PeerNode) bool {
	if p.IP == n.ip && p.Port == n.port {
		return true
	}
	_, isKnownPeer := n.knownPeers[p.Address()]

	return isKnownPeer
}

func (n *Node) mine(ctx context.Context) error {
	var miningCtx context.Context
	var stopCurrentMining context.CancelFunc

	ticker := time.NewTicker(time.Second * mininingIntervalSeconds)

	for {
		select {
		case <-ticker.C:
			go func() {
				if len(n.pendingTRXs) > 0 && !n.isMining {
					n.isMining = true

					miningCtx, stopCurrentMining = context.WithCancel(ctx)

					err := n.minePendingTRXs(miningCtx)
					if err != nil {
						fmt.Println("Error:", err)
					}

					n.isMining = false
				}
			}()
		case block := <-n.newSyncedBlocks:
			if n.isMining {
				blockHash, _ := block.Hash()
				fmt.Println("\nPeer mined next block", blockHash.Hex(), "faster.")

				n.removeMinedPendingTRXs(block)
				stopCurrentMining()
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (n *Node) minePendingTRXs(ctx context.Context) error {
	blockToMine := NewPendingBlock(
		n.state.LatestBlockHash(),
		n.state.LatestBlock().Header.Height+1,
		n.getPendingTRXsAsArray(),
	)

	minedBlock, err := Mine(ctx, blockToMine)
	if err != nil {
		return err
	}

	n.removeMinedPendingTRXs(minedBlock)

	_, err = n.state.AddBlock(minedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) removeMinedPendingTRXs(block db.Block) {
	if len(block.TRXs) > 0 && len(n.pendingTRXs) > 0 {
		fmt.Println("Updating in-memory pending transaction pool:")
	}

	for _, trx := range block.TRXs {
		trxHash, _ := trx.Hash()
		if _, exists := n.pendingTRXs[trxHash.Hex()]; exists {
			fmt.Println("\t-archiving mined transaction:", trxHash.Hex())

			n.archivedTRXs[trxHash.Hex()] = trx
			delete(n.pendingTRXs, trxHash.Hex())
		}
	}
}

func (n *Node) getPendingTRXsAsArray() []db.Trx {
	trxs := make([]db.Trx, 0)
	for _, trx := range n.pendingTRXs {
		trxs = append(trxs, trx)
	}

	return trxs
}

func (pn PeerNode) Address() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}
