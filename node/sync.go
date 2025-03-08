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

func (n *Node) syncBlocks(p PeerNode, status StatusRes) error {
	localBlockHeight := n.state.LatestBlock().Header.Height
	if localBlockHeight < status.Height {
		newBlocksCount := status.Height - localBlockHeight

		fmt.Println("Found", newBlocksCount, "from peer", p.Address())

		blocks, err := fetchBlocksFromPeer(p, n.state.LatestBlockHash())
		if err != nil {
			return err
		}

		if err := n.state.AddBlocks(blocks); err != nil {
			return err
		}

	}
	return nil
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
