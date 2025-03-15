package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	keystoreDirName = "keystore"
	AndrejAccount   = "0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A"
	BabayagaAccount = "0x21973d33e048f5ce006fd7b41f51725c30e4b76b"
	CeasarAccount   = "0x84470a31D271ea400f34e7A697F36bE0e866a716"
)

func NewKeystoreAccount(dataDir, password string) (common.Address, error) {
	ks := keystore.NewKeyStore(GetKeystoreDirPath(dataDir), keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		return common.Address{}, err
	}

	return acc.Address, nil
}

func SignTrxWithKeystoreAccount(
	trx db.Trx, acc common.Address,
	pwd, keystoreDir string,
) (db.SignedTrx, error) {
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	ksAccount, err := ks.Find(accounts.Account{Address: acc})
	if err != nil {
		return db.SignedTrx{}, err
	}

	ksAccountJson, err := fs.AppFS.ReadFile(ksAccount.URL.Path)
	if err != nil {
		return db.SignedTrx{}, err
	}

	key, err := keystore.DecryptKey(ksAccountJson, pwd)
	if err != nil {
		return db.SignedTrx{}, err
	}

	signedTrx, err := SignTrx(trx, key.PrivateKey)
	if err != nil {
		return db.SignedTrx{}, err
	}

	return signedTrx, nil
}

func SignTrx(tx db.Trx, privKey *ecdsa.PrivateKey) (db.SignedTrx, error) {
	rawTx, err := tx.Encode()
	if err != nil {
		return db.SignedTrx{}, err
	}

	sig, err := Sign(rawTx, privKey)
	if err != nil {
		return db.SignedTrx{}, err
	}

	return db.NewSignedTrx(tx, sig), nil
}

func Sign(msg []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	msgHash := sha256.Sum256(msg)

	sig, err = crypto.Sign(msgHash[:], privKey)
	if err != nil {
		return nil, err
	}

	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf("wrong size for signature: got %d want %d", len(sig), crypto.SignatureLength)
	}

	return sig, nil
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	msgHash := sha256.Sum256(msg)

	recoveredPubKey, err := crypto.SigToPub(msgHash[:], sig)
	if err != nil {
		return nil, fmt.Errorf("unable to verify message signature: %s", err.Error())
	}

	return recoveredPubKey, nil
}

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}
