package database

import "strings"

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
