package node

import (
	"context"
	"encoding/hex"
	"testing"

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
	pendingBlock := createRandomPendingBlock()

	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatalf("error mining block: %v", err)
	}

	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatalf("error hashing block: %v", err)
	}

	got := minedBlockHash.IsValid()
	want := true
	if got != want {
		t.Errorf("expected mined block hash to be valid")
	}
}

func createRandomPendingBlock() PendingBlock {
	return NewPendingBlock(db.Hash{}, 0, []db.Trx{
		db.NewTrx("andrej", "andrej", 3, ""),
		db.NewTrx("andrej", "andrej", 700, ""),
	})
}
