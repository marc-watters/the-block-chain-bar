package database

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type (
	Account string

	Trx struct {
		From  common.Address `json:"from"`
		To    common.Address `json:"to"`
		Value uint64         `json:"value"`
		Data  string         `json:"data"`
		Time  uint64         `json:"time"`
	}
	SignedTrx struct {
		Trx
		Sig []byte `json:"signature"`
	}
)

func NewAccount(value string) common.Address {
	return common.HexToAddress(value)
}

func NewSignedTrx(trx Trx, sig []byte) SignedTrx {
	return SignedTrx{trx, sig}
}

func NewTrx(from common.Address, to common.Address, value uint64, data string) Trx {
	return Trx{from, to, value, data, uint64(time.Now().UnixNano())}
}

func (t Trx) IsReward() bool {
	return t.Data == "reward"
}

func (t Trx) Hash() (Hash, error) {
	trxJSON, err := t.Encode()
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(trxJSON), nil
}

func (st SignedTrx) Hash() (Hash, error) {
	trxJSON, err := st.Encode()
	if err != nil {
		return Hash{}, nil
	}
	return sha256.Sum256(trxJSON), nil
}

func (st SignedTrx) IsAuthentic() (bool, error) {
	trxHash, err := st.Trx.Hash()
	if err != nil {
		return false, err
	}

	recoveredPubKey, err := crypto.SigToPub(trxHash[:], st.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	return recoveredAccount.Hex() == st.From.Hex(), nil
}
