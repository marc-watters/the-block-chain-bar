package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var AppFs *afero.Afero

type State struct {
	Balances map[Account]uint

	txMempool []Tx
	db        afero.File
}

const (
	Dir  = "database"
	GenF = "genesis.json"
	TxF  = "tx.db"
)

func NewStateFromDisk() (*State, error) {
	// load genesis file
	g, err := loadGenesis(filepath.Join(Dir, GenF))
	if err != nil {
		return nil, err
	}

	// load transaction file
	txf, err := AppFs.OpenFile(filepath.Join(Dir, TxF), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// create state object
	s := &State{
		Balances:  make(map[Account]uint),
		txMempool: make([]Tx, 0),
		db:        txf,
	}

	// populate state balances
	for account, balance := range g.Balances {
		s.Balances[account] = balance
	}

	// process transactions
	if _, err := s.db.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(s.db)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("tx scan failed: %v", err)
		}

		var tx Tx
		if err := json.Unmarshal(scanner.Bytes(), &tx); err != nil {
			return nil, fmt.Errorf("unmarshall transaction: %v", err)
		}

		if tx.IsReward() {
			s.Balances[tx.To] += tx.Value
			continue
		}

		if s.Balances[tx.From] < tx.Value {
			return nil, fmt.Errorf("insufficient balance")
		}

		s.Balances[tx.From] -= tx.Value
		s.Balances[tx.To] += tx.Value
	}

	return s, nil
}
