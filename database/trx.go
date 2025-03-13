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
)

func NewAccount(value string) common.Address {
	return common.HexToAddress(value)
}

func NewTrx(from common.Address, to common.Address, value uint64, data string) Trx {
	return Trx{from, to, value, data, uint64(time.Now().UnixNano())}
}

func (t Trx) IsReward() bool {
	return t.Data == "reward"
}

func (t Trx) Hash() (Hash, error) {
	trxJSON, err := json.Marshal(t)
	if err != nil {
		return Hash{}, nil
	}
	return sha256.Sum256(trxJSON), nil
}
