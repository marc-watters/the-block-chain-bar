package database

import (
	"crypto/sha256"
	"encoding/json"
	"strings"
)

type (
	Account string

	Trx struct {
		From  Account
		To    Account
		Value uint64
		Data  string
	}
)

func NewAccount(value string) Account {
	return Account(strings.ToLower(value))
}

func NewTrx(from Account, to Account, value uint64, data string) Trx {
	return Trx{from, to, value, data}
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
