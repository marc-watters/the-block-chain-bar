package node

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

func TestValidBlockHash(t *testing.T) {
	hexHash := "000000fa04f8160395c387277f8b2f14837603383d33809a4db586086168edfa"

	var hash db.Hash
	if _, err := hex.Decode(hash[:], []byte(hexHash)); err != nil {
		t.Fatalf("error decoding hash: %v", err)
	}

	got := hash.IsValid()
	want := true

	if got != want {
		t.Errorf("hash should be valid: %s:", hexHash)
	}
}

func TestInvalidBlockHash(t *testing.T) {
	hexHash := "000001fa04f8160395c387277f8b2f14837603383d33809a4db586086168edfa"

	var hash db.Hash
	if _, err := hex.Decode(hash[:], []byte(hexHash)); err != nil {
		t.Fatalf("error decoding hash: %v", err)
	}

	got := hash.IsValid()
	want := false
	if got != want {
		t.Errorf("hash should be invalid: %s:", hexHash)
	}
}

func TestMine(t *testing.T) {
	miner := db.NewAccount("andrej")
	pendingBlock := createRandomPendingBlock(miner)

	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatalf("error mining block: %v", err)
	}

	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatalf("error hashing block: %v", err)
	}

	if !minedBlockHash.IsValid() {
		t.Fatal("expected mined block hash to be valid")
	}

	if minedBlock.Header.Miner != miner {
		t.Errorf("mined block miner should equal miner from pending block")
	}
}

func TestMineWithTimeout(t *testing.T) {
	miner := db.NewAccount("andrej")
	pendingBlock := createRandomPendingBlock(miner)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)

	if _, err := Mine(ctx, pendingBlock); err == nil {
		t.Errorf("expected timeout error mining block: %v", err)
		cancel()
	}
	cancel()
}

func createRandomPendingBlock(miner db.Account) PendingBlock {
	return NewPendingBlock(
		db.Hash{},
		1,
		miner,
		[]db.Trx{
			db.NewTrx("andrej", "andrej", 3, ""),
			db.NewTrx("andrej", "andrej", 700, ""),
		})
}
