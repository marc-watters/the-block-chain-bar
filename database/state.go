package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var AppFS *afero.Afero

func init() {
	AppFS = &afero.Afero{Fs: afero.NewOsFs()}
}

type State struct {
	Balances map[Account]uint64

	trxMempool []Trx
	db         afero.File
}

func NewStateFromDisk() (*State, error) {
	s := &State{
		Balances:   make(map[Account]uint64),
		trxMempool: make([]Trx, 0),
		db:         nil,
	}

	g, err := loadGenesis(filepath.Join("database", "genesis.json"))
	if err != nil {
		return nil, err
	}
	maps.Copy(s.Balances, g.Balances)

	s.db, err = AppFS.OpenFile(filepath.Join("database", "trx.db"), os.O_APPEND|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(s.db)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var trx Trx
		if err := json.Unmarshal(scanner.Bytes(), &trx); err != nil {
			return nil, err
		}

		if trx.Data == "reward" {
			s.Balances[trx.To] += trx.Value
			continue
		}

		if trx.Value > s.Balances[trx.From] {
			return nil, fmt.Errorf("insufficient balance")
		}

		s.Balances[trx.From] -= trx.Value
		s.Balances[trx.To] += trx.Value
	}

	return s, nil
}
