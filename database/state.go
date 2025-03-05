package database

import (
	"bufio"
	"crypto/sha256"
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

const (
	Dir     = "database"
	GenFile = "genesis.json"
	TrxFile = "trx.db"
)

type (
	Snapshot [32]byte

	State struct {
		Balances        map[Account]uint64
		latestBlockHash Hash
		trxMempool      []Trx
		db              afero.File
	}
)

func NewStateFromDisk() (*State, error) {
	s := &State{
		Balances:        make(map[Account]uint64),
		latestBlockHash: Hash{},
		trxMempool:      make([]Trx, 0),
		db:              nil,
	}

	g, err := loadGenesis(filepath.Join(Dir, GenFile))
	if err != nil {
		return nil, err
	}
	maps.Copy(s.Balances, g.Balances)

	s.db, err = AppFS.OpenFile(filepath.Join(Dir, TrxFile), os.O_APPEND|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if _, err = s.db.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(s.db)
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, err
		}

		var trx Trx
		if err = json.Unmarshal(scanner.Bytes(), &trx); err != nil {
			return nil, err
		}

		if err = s.apply(trx); err != nil {
			return nil, err
		}
	}

	err = s.doSnapshot()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *State) Add(trx Trx) error {
	if err := s.apply(trx); err != nil {
		return err
	}
	s.trxMempool = append(s.trxMempool, trx)

	return nil
}

func (s *State) Persist() (Snapshot, error) {
	for len(s.trxMempool) > 0 {
		var trx Trx
		trx, s.trxMempool = s.trxMempool[0], s.trxMempool[1:]

		trxJSON, err := json.Marshal(trx)
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Println("Persisting new transaction to disk:")
		fmt.Printf("\t%s", trxJSON)
		fmt.Println()
		_, err = s.db.Write(append(trxJSON, '\n'))
		if err != nil {
			return Snapshot{}, err
		}

		err = s.doSnapshot()
		if err != nil {
			return Snapshot{}, err
		}
		fmt.Printf("New DB Snapshot: %x\n", s.snapshot)
	}

	return s.snapshot, nil
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() error {
	return s.db.Close()
}

func (s *State) apply(trx Trx) error {
	if trx.From == "" {
		return NewInvalidTransaction("From")
	}
	if trx.To == "" {
		return NewInvalidTransaction("To")
	}
	if trx.Value == 0 {
		return NewInvalidTransaction("Value")
	}

	if trx.IsReward() {
		s.Balances[trx.To] += trx.Value
		return nil
	}

	if trx.Value > s.Balances[trx.From] {
		return new(ErrInsufficientBalance)
	}

	s.Balances[trx.From] -= trx.Value
	s.Balances[trx.To] += trx.Value

	return nil
}

func (s *State) doSnapshot() error {
	_, err := s.db.Seek(0, 0)
	if err != nil {
		return err
	}

	trxsData, err := io.ReadAll(s.db)
	if err != nil {
		return err
	}

	s.snapshot = sha256.Sum256(trxsData)

	return nil
}
