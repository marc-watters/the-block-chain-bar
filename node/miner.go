package node

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type PendingBlock struct {
	parent db.Hash
	height uint64
	time   uint64
	miner  common.Address
	trxs   []db.Trx
}

func NewPendingBlock(parent db.Hash, height uint64, miner common.Address, trxs []db.Trx) PendingBlock {
	t := uint64(time.Now().UnixNano())
	return PendingBlock{parent, height, t, miner, trxs}
}

func Mine(ctx context.Context, pb PendingBlock) (db.Block, error) {
	if len(pb.trxs) == 0 {
		return db.Block{}, fmt.Errorf("mining empty blocks is not allowed")
	}

	sort.Slice(pb.trxs, func(i, j int) bool {
		return pb.trxs[i].Time < pb.trxs[j].Time
	})

	start := time.Now()
	attempt := 0
	var (
		block db.Block
		hash  db.Hash
		nonce uint32
	)

	for !hash.IsValid() {
		select {
		case <-ctx.Done():
			fmt.Println("Mining cancelled")
			return db.Block{}, ctx.Err()
		default:
		}

		attempt++
		nonce = generateNonce()

		if attempt%1000000 == 0 || attempt == 1 {
			fmt.Println("Mining", len(pb.trxs), "pending transactions. Attempt:", attempt)
		}

		block = db.NewBlock(pb.parent, pb.height, nonce, pb.time, pb.miner, pb.trxs)
		blockHash, err := block.Hash()
		if err != nil {
			return db.Block{}, fmt.Errorf("couldn't mine block: %v", err)
		}

		hash = blockHash
	}

	fmt.Printf("\nMined new Block '%x' using PoW🎉🎉🎉:\n", hash)
	fmt.Printf("\tHeight: '%v'\n", block.Header.Height)
	fmt.Printf("\tNonce: '%v'\n", block.Header.Nonce)
	fmt.Printf("\tCreated: '%v'\n", block.Header.Time)
	fmt.Printf("\tMiner '%v'\n", block.Header.Miner.String())
	fmt.Printf("\tParent: '%v'\n\n", block.Header.Parent.Hex())

	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(start))

	return block, nil
}

func generateNonce() uint32 {
	return rand.Uint32()
}
