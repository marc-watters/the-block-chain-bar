package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

// The password for testing keystore files:
//
//	./node/test_andrej--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57
//	./node/test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8
const testKeystoreAccountsPwd = "security123"

func TestSign(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubKey := privKey.PublicKey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	pubKeyBytesHash := crypto.Keccak256(pubKeyBytes[:])

	account := common.BytesToAddress(pubKeyBytesHash[1:])

	msg := []byte("the Web3Coach students are awesome")

	sig, err := Sign(msg, privKey)
	if err != nil {
		t.Fatal(err)
	}

	recoveredPubKey, err := Verify(msg, sig)
	if err != nil {
		t.Fatal(err)
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	if account.Hex() != recoveredAccount.Hex() {
		t.Fatalf("msg was signed by account %s but signature reecovery produced account %s", account.Hex(), recoveredAccount.Hex())
	}
}

func TestSignTrxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := fs.AppFS.TempDir("", "wallet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fs.RemoveDir(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "error removing data directory: %v", err)
		}
	}()

	andrej, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	babaYaga, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	tx := db.NewTrx(andrej, babaYaga, 100, "")

	signedTrx, err := SignTrxWithKeystoreAccount(tx, andrej, testKeystoreAccountsPwd, GetKeystoreDirPath(tmpDir))
	if err != nil {
		t.Error(err)
		return
	}

	spew.Dump(signedTrx.Encode())
	ok, err := signedTrx.IsAuthentic()
	if err != nil {
		t.Error(err)
		return
	}

	if !ok {
		t.Fatal("the transaction was signed by 'from' account and should have been authentic")
	}
}

func TestSignForgedTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := fs.AppFS.TempDir("", "wallet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fs.RemoveDir(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "error removing data directory: %v", err)
		}
	}()

	hacker, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	babaYaga, err := NewKeystoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil {
		t.Error(err)
		return
	}

	forgedTrx := db.NewTrx(babaYaga, hacker, 100, "")

	signedTrx, err := SignTrxWithKeystoreAccount(forgedTrx, hacker, testKeystoreAccountsPwd, GetKeystoreDirPath(tmpDir))
	if err != nil {
		t.Error(err)
		return
	}

	ok, err := signedTrx.IsAuthentic()
	if err != nil {
		t.Error(err)
		return
	}

	if ok {
		t.Fatal("the transaction 'from' attribute was forged and should have not be authentic")
	}
}
