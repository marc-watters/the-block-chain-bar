package database

import (
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"
)

type (
	Account string

	Trx struct {
		From  Account `json:"from"`
		To    Account `json:"to"`
		Value uint64  `json:"value"`
		Data  string  `json:"data"`
		Time  uint64  `json:"time"`
	}
)

func NewAccount(value string) Account {
	return Account(strings.ToLower(value))
}

func NewTrx(from Account, to Account, value uint64, data string) Trx {
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
