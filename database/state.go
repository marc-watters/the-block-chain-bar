package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"time"

	"github.com/spf13/afero"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

type State struct {
	balances        map[Account]uint64
	latestBlock     Block
	latestBlockHash Hash
	trxMempool      []Trx
	dataDir         string
	db              afero.File
}

func NewStateFromDisk(dataDir string) (*State, error) {
	dataDir = fs.ExpandPath(dataDir)

	err := fs.InitDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	s := &State{
		balances:        make(map[Account]uint64),
		latestBlock:     Block{},
		latestBlockHash: Hash{},
		trxMempool:      make([]Trx, 0),
		dataDir:         dataDir,
		db:              nil,
	}

	g, err := loadGenesis(fs.GetGenesisJSONFilePath(dataDir))
	if err != nil {
		return nil, err
	}
	maps.Copy(s.balances, g.Balances)

	s.db, err = fs.AppFS.OpenFile(
		fs.GetBlocksDBFilePath(dataDir),
		os.O_APPEND|os.O_RDWR,
		0o600,
	)
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

		if len(blockFSJSON) == 0 {
			break
		}

		err = s.applyBlock(blockFS.Value)
		if err != nil {
			return nil, err
		}

		s.latestBlockHash = blockFS.Key
	}

	return s, nil
}

func (s *State) DataDir() string {
	return s.dataDir
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
	latestBlockHash, err := s.latestBlock.Hash()
	if err != nil {
		return Hash{}, nil
	}

	block := NewBlock(
		s.latestBlockHash,
		s.latestBlock.Header.Height+1,
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
	s.latestBlock = block
	s.trxMempool = []Trx{}

	return latestBlockHash, nil
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Balances() map[Account]uint64 {
	return s.balances
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
		s.balances[trx.To] += trx.Value
		return nil
	}

	if trx.Value > s.balances[trx.From] {
		return new(ErrInsufficientBalance)
	}

	s.balances[trx.From] -= trx.Value
	s.balances[trx.To] += trx.Value

	return nil
}
