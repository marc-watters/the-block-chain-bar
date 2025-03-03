package database

import "github.com/spf13/afero"

type State struct {
	Balances map[string]uint64

	trxMempool []struct {
		From  string
		To    string
		Value uint64
		Data  string
	}

	db afero.File
}
