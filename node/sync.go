package node

import (
	"context"
	"fmt"
	"net/http"
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(45 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Searching for new Peers and Blocks...")

			n.doSync()

		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) doSync() {
	for _, peer := range n.knownPeers {
		if n.ip == peer.IP && n.port == peer.Port {
			continue
		}

		fmt.Println("Querying new peers and blocks from:", peer.Address())
		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Println("ERROR:", err)
			fmt.Println("Peer", peer.Address(), "was removed from known peers")
			n.deletePeer(peer)
			continue
		}

		if err := n.joinKnownPeers(peer); err != nil {
			fmt.Println("ERROR:", err)
			continue
		}

		if err := n.syncBlocks(peer, status); err != nil {
			fmt.Println("ERROR:", err)
			continue
		}

		if err := n.syncKnownPeers(status); err != nil {
			fmt.Println("ERROR:", err)
			continue
		}

	}
}

func (n *Node) syncBlocks(p PeerNode, status StatusRes) error {
	localBlockHeight := n.state.LatestBlock().Header.Height

	if status.Hash.IsEmpty() {
		return nil
	}

	if status.Height < localBlockHeight {
		return nil
	}

	if status.Height == 0 && !n.state.LatestBlockHash().IsEmpty() {
		return nil
	}

	newBlocksCount := status.Height - localBlockHeight
	if localBlockHeight == 0 && status.Height == 0 {
		newBlocksCount = 1
	}

	fmt.Println("Found", newBlocksCount, "new blocks from peer:", p.Address())

	if newBlocksCount == 0 {
		return nil
	}

	blocks, err := fetchBlocksFromPeer(p, n.state.LatestBlockHash())
	if err != nil {
		return err
	}

	return n.state.AddBlocks(blocks)
}

func (n *Node) syncKnownPeers(status StatusRes) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.isKnownPeer(statusPeer) {
			fmt.Println("Found new peer:", statusPeer.Address())
			n.addPeer(statusPeer)
		}
	}

	return nil
}

func (n *Node) syncPendingTRXs(p PeerNode, status StatusRes) error {
	for _, trx := range status.PendingTRXs {
		if err := n.AddPendingTrx(trx, p); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) joinKnownPeers(p PeerNode) error {
	if p.connected {
		return nil
	}

	hostpath := "http://%s%s"
	queryIP := "?%s=%s&"
	queryPort := "%s=%d"
	url := fmt.Sprintf("%s%s%s",
		fmt.Sprintf(hostpath, p.Address(), endpointAddPeer),
		fmt.Sprintf(queryIP, endpointAddPeerQueryKeyIP, n.ip),
		fmt.Sprintf(queryPort, endpointAddPeerQueryKeyPort, n.port),
	)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	var addPeerRes AddPeerRes
	if err := readRes(res, &addPeerRes); err != nil {
		return err
	}

	if addPeerRes.Error != "" {
		return fmt.Errorf("error: %v", addPeerRes.Error)
	}

	knownPeer := n.knownPeers[p.Address()]
	knownPeer.connected = addPeerRes.Success

	n.addPeer(knownPeer)

	if !addPeerRes.Success {
		return fmt.Errorf("unable to join known peers of '%s'", p.Address())
	}

	return nil
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s/%s", peer.Address(), endpointStatus)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, err
	}

	var statusRes StatusRes
	err = readRes(res, &statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}

func fetchBlocksFromPeer(p PeerNode, fromBlock db.Hash) ([]db.Block, error) {
	fmt.Println("Importing blocks...")

	url := fmt.Sprintf(
		"http://%s%s?%s=%s",
		p.Address(),
		endpointSync,
		endpointSyncQueryKeyFromBlock,
		fromBlock.Hex(),
	)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var syncRes SyncRes
	err = readRes(res, &syncRes)
	if err != nil {
		return nil, err
	}

	return syncRes.Blocks, nil
}
