package node

import (
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type PendingBlock struct {
	parent db.Hash
	height uint64
	time   uint64
	trxs   []db.Trx
}

func NewPendingBlock(parent db.Hash, height uint64, trxs []db.Trx) PendingBlock {
	t := uint64(time.Now().Unix())
	return PendingBlock{parent, height, t, trxs}
}

func generateNonce() uint32 {
	return rand.Uint32()
}
