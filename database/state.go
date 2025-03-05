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
	"time"

	"github.com/spf13/afero"
)

var AppFS *afero.Afero

func init() {
	AppFS = &afero.Afero{Fs: afero.NewOsFs()}
}

const (
	Dir     = "database"
	GenFile = "genesis.json"
	TrxFile = "block.db"
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

		var blockFS BlockFS
		blockFSJSON := scanner.Bytes()
		err = json.Unmarshal(blockFSJSON, &blockFS)
		if err != nil {
			return nil, err
		}

		err = s.applyBlock(blockFS.Value)
		if err != nil {
			return nil, err
		}

		s.latestBlockHash = blockFS.Key
	}

	return s, nil
}

func (s *State) AddBlock(b Block) error {
	for _, trx := range b.TRXs {
		if err := s.AddTrx(trx); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) AddTrx(trx Trx) error {
	if err := s.apply(trx); err != nil {
		return err
	}
	s.trxMempool = append(s.trxMempool, trx)

	return nil
}

func (s *State) Persist() (Hash, error) {
	block := NewBlock(
		s.latestBlockHash,
		uint64(time.Now().Unix()),
		s.trxMempool,
	)

	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFS := BlockFS{blockHash, block}

	blockFSJON, err := json.Marshal(blockFS)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println()
	fmt.Println("Persisting new Block to disk:")
	fmt.Printf("\t%s", blockFSJON)
	fmt.Println()

	if _, err := s.db.Write(append(blockFSJON, '\n')); err != nil {
		return Hash{}, err
	}

	s.latestBlockHash = blockHash

	return s.latestBlockHash, nil
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() error {
	return s.db.Close()
}

func (s *State) applyBlock(b Block) error {
	for _, trx := range b.TRXs {
		if err := s.apply(trx); err != nil {
			return err
		}
	}
	return nil
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
