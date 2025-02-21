package database

import "github.com/spf13/afero"

type State struct {
	Balances map[Account]uint

	txMempool []Tx
	db        *afero.File
}
