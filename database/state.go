package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
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

		if err := applyTRXs(blockFS.Value.TRXs, s); err != nil {
			return nil, err
		}

		s.latestBlock = blockFS.Value
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
	if err := applyTrx(trx, s); err != nil {
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

func applyBlock(b Block, s State) error {
	nextExpectedBlockHeight := s.latestBlock.Header.Height + 1

	if b.Header.Height != nextExpectedBlockHeight {
		return fmt.Errorf("next expected block height must be '%d' not '%d'",
			nextExpectedBlockHeight,
			b.Header.Height,
		)
	}

	if s.latestBlock.Header.Height > 0 && !reflect.DeepEqual(
		b.Header.Parent, s.latestBlockHash,
	) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'",
			s.latestBlockHash, b.Header.Parent,
		)
	}

	return applyTRXs(b.TRXs, &s)
}

func applyTrx(trx Trx, s *State) error {
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

func applyTRXs(trxs []Trx, s *State) error {
	for _, trx := range trxs {
		err := applyTrx(trx, s)
		if err != nil {
			return err
		}
	}

	return nil
}
