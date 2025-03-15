package node

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
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
	minerPrivKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivKey, miner)
	if err != nil {
		t.Fatal(err)
	}

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
	minerPrivKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivKey, miner)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()

	if _, err := Mine(ctx, pendingBlock); err == nil {
		t.Errorf("expected timeout error mining block: %v", err)
	}
}

func generateKey() (*ecdsa.PrivateKey, ecdsa.PublicKey, common.Address, error) {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, ecdsa.PublicKey{}, common.Address{}, err
	}

	pubKey := privKey.PublicKey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	pubKeyBytesHash := crypto.Keccak256(pubKeyBytes[1:])

	account := common.BytesToAddress(pubKeyBytesHash[12:])

	return privKey, pubKey, account, nil
}

func createRandomPendingBlock(privKey *ecdsa.PrivateKey, acc common.Address) (PendingBlock, error) {
	trx := db.NewTrx(acc, db.NewAccount(testKsBabaYagaAccount), 1, "")
	signedTrx, err := wallet.SignTrx(trx, privKey)
	if err != nil {
		return PendingBlock{}, err
	}

	return NewPendingBlock(
		db.Hash{},
		0, acc,
		[]db.SignedTrx{signedTrx},
	), nil
}
