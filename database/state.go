package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/spf13/afero"
)

var AppFs *afero.Afero

type State struct {
	Balances map[Account]uint

	txMempool []Tx
	db        afero.File

	latestBlock     Block
	latestBlockHash Hash
	hasGenesisBlock bool
}

const (
	Dir  = "database"
	GenF = "genesis.json"
	TxF  = "block.db"
)

func NewStateFromDisk(dataDir string) (*State, error) {
	dataDir = ExpandPath(dataDir)

	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	// load genesis file
	g, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	// load transaction file
	txf, err := AppFs.OpenFile(getBlocksDbFilePath(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// create state object
	s := &State{
		Balances:        make(map[Account]uint),
		txMempool:       make([]Tx, 0),
		latestBlock:     Block{},
		latestBlockHash: Hash{},
		hasGenesisBlock: false,
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

		if len(bfsJson) == 0 {
			break
		}

		var bfs BlockFS
		if err := json.Unmarshal(bfsJson, &bfs); err != nil {
			return nil, err
		}

		if err := s.applyBlock(bfs.Value); err != nil {
			return nil, err
		}

		s.latestBlock = bfs.Value
		s.latestBlockHash = bfs.Key
		s.hasGenesisBlock = true
	}

	return s, nil
}

func (s *State) AddTx(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	err := applyBlock(b, pendingState)
	if err != nil {
		return Hash{}, err
	}

	bh, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	bfs := BlockFS{bh, b}

	bfsJson, err := json.Marshal(bfs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println("Persisting new block to disk:")
	fmt.Printf("\t%s\t\n", bfsJson)

	_, err = s.db.Write(append(bfsJson, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = pendingState.Balances
	s.hasGenesisBlock = true
	s.latestBlockHash = bh
	s.latestBlock = b

	return bh, nil
}

func (s *State) Persist() (Hash, error) {
	lbh, err := s.latestBlock.Hash()
	if err != nil {
		return Hash{}, err
	}

	b := NewBlock(
		lbh,
		s.latestBlock.Header.Height+1,
		uint64(time.Now().Unix()),
		s.txMempool,
	)

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

	s.latestBlockHash = lbh
	s.latestBlock = b

	s.txMempool = []Tx{}

	return lbh, nil
}

func (s *State) LatestHash() Hash {
	return s.latestBlockHash
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) NextBlockHeight() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LatestBlock().Header.Height + 1
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

func (s *State) copy() State {
	c := State{}
	c.hasGenesisBlock = s.hasGenesisBlock
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.txMempool = make([]Tx, len(s.txMempool))
	c.Balances = make(map[Account]uint)

	for acc, bal := range s.Balances {
		c.Balances[acc] = bal
	}

	c.txMempool = append(c.txMempool, s.txMempool...)

	return c
}

func applyBlock(b Block, s State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Height + 1

	if s.hasGenesisBlock && b.Header.Height != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be %q not %q", nextExpectedBlockNumber, b.Header.Height)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Height > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' '%x'", s.latestBlockHash, b.Header.Parent)
	}

	return applyTXs(b.Payload, &s)
}

func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyTx(tx Tx, s *State) error {
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
