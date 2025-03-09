package node

import (
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
