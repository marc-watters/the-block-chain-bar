package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

var AppFs *afero.Afero

type State struct {
	Balances map[Account]uint

	txMempool []Tx
	db        afero.File

	latestBlockHash Hash
}

const (
	Dir  = "database"
	GenF = "genesis.json"
	TxF  = "block.db"
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
		Balances:        make(map[Account]uint),
		txMempool:       make([]Tx, 0),
		latestBlockHash: Hash{},
		db:              txf,
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

		bfsJson := scanner.Bytes()
		var bfs BlockFS
		if err := json.Unmarshal(bfsJson, &bfs); err != nil {
			return nil, err
		}

		if err := s.applyBlock(bfs.Value); err != nil {
			return nil, err
		}

		s.latestBlockHash = bfs.Key
	}

	return s, nil
}

func (s *State) Add(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) AddBlock(b Block) error {
	for _, tx := range b.Payload {
		if err := s.AddTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) Persist() (Hash, error) {
	b := NewBlock(s.latestBlockHash, uint64(time.Now().Unix()), s.txMempool)
	bh, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	bfs := BlockFS{bh, b}

	bfsJson, err := json.Marshal(bfs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", bfsJson)

	if _, err := s.db.Write(append(bfsJson, '\n')); err != nil {
		return Hash{}, err
	}

	s.latestBlockHash = bh

	s.txMempool = []Tx{}

	return s.latestBlockHash, nil
}

func (s *State) LatestHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() {
	s.db.Close()
}

func (s *State) applyBlock(b Block) error {
	for _, tx := range b.Payload {
		if err := s.apply(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return ErrInsufficientBalance{tx.From, tx.To, tx.Value}
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}
