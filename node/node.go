package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
)

const (
	DefaultBootstrapIP   = "node.tbb.web3.coach"
	DefaultBootstrapPort = 8080
	DefaultBootstrapAcc  = wallet.AndrejAccount
	DefaultMiner         = "0x0000000000000000000000000000000000000000"

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
	endpointAddPeerQueryKeyMiner  = "miner"

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
		LatestBlock() db.Block
		LatestBlockHash() db.Hash
		NextBlockHeight() uint64
		Balances() map[common.Address]uint64
		DataDir() string
	}
	PeerNode struct {
		IP         string         `json:"ip"`
		Port       uint64         `json:"port"`
		IsBoostrap bool           `json:"is_bootstrap"`
		Account    common.Address `json:"account"`

		connected bool
	}
)

func New(s state, ip string, port uint64, acc common.Address, bootstrap PeerNode) *Node {
	knownPeers := map[string]PeerNode{
		bootstrap.Address(): bootstrap,
	}
	return &Node{
		info:  NewPeerNode(ip, port, false, acc, true),
		state: s,

		knownPeers:      knownPeers,
		pendingTRXs:     make(map[string]db.Trx),
		archivedTRXs:    make(map[string]db.Trx),
		newSyncedBlocks: make(chan db.Block),
		newPendingTRXs:  make(chan db.Trx, 10000),
		isMining:        false,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, acc common.Address, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, acc, connected}
}

func (n *Node) Run(ctx context.Context) error {
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
	go func() {
		if err := n.mine(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "Node.Run() mine error:", err)
		}
	}()

	fmt.Printf("Listening on %s:%d\n", n.info.IP, n.info.Port)

	fmt.Println("Blockchain state:")
	fmt.Printf("	- height: %d\n", n.state.LatestBlock().Header.Height)
	fmt.Printf("	- hash: %s\n", n.state.LatestBlockHash().Hex())

	server := &http.Server{Addr: fmt.Sprintf(":%d", n.info.Port)}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (n *Node) LatestBlockHash() db.Hash {
	return n.state.LatestBlockHash()
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

	if err := n.AddPendingTrx(trx, n.info); err != nil {
		writeErr(w, err)
		return
	}

	writeRes(w, TrxPostRes{Success: true})
}

func (n *Node) Status(w http.ResponseWriter, r *http.Request) {
	res := StatusRes{
		Hash:        n.state.LatestBlockHash(),
		Height:      n.state.LatestBlock().Header.Height,
		KnownPeers:  n.knownPeers,
		PendingTRXs: n.getPendingTRXsAsArray(),
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

	minerRaw := r.URL.Query().Get(endpointAddPeerQueryKeyPort)

	peerPort, err := strconv.ParseUint(peerPortRaw, 10, 32)
	if err != nil {
		writeRes(w, AddPeerRes{false, err.Error()})
		return
	}

	peer := NewPeerNode(peerIP, peerPort, false, db.NewAccount(minerRaw), true)

	n.addPeer(peer)

	fmt.Println("Peer", peer.Address(), "was added to known peers")

	writeRes(w, AddPeerRes{true, ""})
}

func (n *Node) AddPendingTrx(trx db.Trx, fromPeer PeerNode) error {
	trxHash, err := trx.Hash()
	if err != nil {
		return err
	}

	trxJSON, err := json.Marshal(trx)
	if err != nil {
		return err
	}

	_, isAlreadyPending := n.pendingTRXs[trxHash.Hex()]
	_, isArchived := n.archivedTRXs[trxHash.Hex()]

	if !isAlreadyPending && !isArchived {
		fmt.Printf("[%s]- added pending transaction %s from peer %s\n", n.info.Address(), trxJSON, fromPeer.Address())
		n.pendingTRXs[trxHash.Hex()] = trx
		n.newPendingTRXs <- trx
	}

	return nil
}

func (n *Node) addPeer(p PeerNode) {
	n.knownPeers[p.Address()] = p
}

func (n *Node) deletePeer(p PeerNode) {
	delete(n.knownPeers, p.Address())
}

func (n *Node) isKnownPeer(p PeerNode) bool {
	if p.IP == n.info.IP && p.Port == n.info.Port {
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

					if err := n.minePendingTRXs(miningCtx); err != nil {
						if !errors.Is(err, context.Canceled) {
							fmt.Println("ERROR:", err)
							fmt.Println("error while mining:", err)
						}
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
		n.info.Account,
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
	trxs := make([]db.Trx, len(n.pendingTRXs))

	var i int
	for _, trx := range n.pendingTRXs {
		trxs[i] = trx
		i++
	}

	return trxs
}

func (pn PeerNode) Address() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}
