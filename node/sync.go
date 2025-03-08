package node

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(45 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Searching for new Peers and Blocks...")

			n.fetchNewBlocksAndPeers()

		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) fetchNewBlocksAndPeers() {
	for _, knownPeer := range n.knownPeers {
		status, err := queryPeerStatus(knownPeer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}

		localBlockHeight := n.state.LatestBlock().Header.Height
		if localBlockHeight < status.Height {
			newBlocksCount := status.Height - localBlockHeight

			fmt.Printf("Found %d new blocks from Peer %s\n", newBlocksCount, knownPeer.IP)
		}

		for _, peer := range status.KnownPeers {
			newPeer, isKnownPeer := n.knownPeers[peer.Address()]
			if !isKnownPeer {
				fmt.Printf("Found new Peer %s\n", newPeer.Address())

				n.knownPeers[newPeer.Address()] = newPeer
			}
		}
	}
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
