package database

import "strings"

type Account string

func NewAccount(value string) Account {
	return Account(strings.ToLower(value))
}

type Trx struct {
	From  Account
	To    Account
	Value uint64
	Data  string
}

func NewTrx(from Account, to Account, value uint64, data string) Trx {
	return Trx{from, to, value, data}
}
