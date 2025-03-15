package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/afero"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

type State struct {
	balances        map[common.Address]uint64
	latestBlock     Block
	latestBlockHash Hash
	hasGenesisBlock bool
	dataDir         string
	db              afero.File
}

func NewStateFromDisk(dataDir string) (*State, error) {
	dataDir = fs.ExpandPath(dataDir)

	err := fs.InitDataDirIfNotExists(dataDir, []byte(fs.GenesisJSON))
	if err != nil {
		return nil, err
	}

	s := &State{
		balances:        make(map[common.Address]uint64),
		latestBlock:     Block{},
		latestBlockHash: Hash{},
		hasGenesisBlock: false,
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

		if err := applyBlock(blockFS.Value, s); err != nil {
			return nil, err
		}

		s.latestBlock = blockFS.Value
		s.latestBlockHash = blockFS.Key
		s.hasGenesisBlock = true
	}

	return s, nil
}

func (s *State) DataDir() string {
	return s.dataDir
}

func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		if _, err := s.AddBlock(b); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	if err := applyBlock(b, &pendingState); err != nil {
		return Hash{}, err
	}

	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFS := BlockFS{blockHash, b}

	blockFSJSON, err := json.Marshal(blockFS)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println("Persisting new Block to disk:")
	fmt.Printf("\t%s\n", blockFSJSON)
	if _, err := s.db.Write(append(blockFSJSON, '\n')); err != nil {
		return Hash{}, err
	}

	s.balances = pendingState.balances
	s.latestBlockHash = blockHash
	s.latestBlock = b
	s.hasGenesisBlock = true

	return blockHash, nil
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) NextBlockHeight() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.latestBlock.Header.Height + 1
}

func (s *State) Balances() map[common.Address]uint64 {
	return s.balances
}

func (s *State) Close() error {
	return s.db.Close()
}

func (s *State) copy() State {
	c := State{}
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.hasGenesisBlock = s.hasGenesisBlock
	c.balances = make(map[common.Address]uint64)

	maps.Copy(c.balances, s.balances)

	return c
}

func applyBlock(b Block, s *State) error {
	if s.hasGenesisBlock && b.Header.Height != s.NextBlockHeight() {
		return fmt.Errorf("next expected block height must be '%d' not '%d'",
			s.NextBlockHeight(),
			b.Header.Height,
		)
	}

	if s.hasGenesisBlock &&
		s.latestBlock.Header.Height > 0 &&
		!reflect.DeepEqual(
			b.Header.Parent, s.latestBlockHash,
		) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'",
			s.latestBlockHash, b.Header.Parent,
		)
	}

	hash, err := b.Hash()
	if err != nil {
		return err
	}

	if !hash.IsValid() {
		return fmt.Errorf("invalid block hash %x", hash)
	}

	if err := applyTRXs(b.TRXs, s); err != nil {
		return err
	}

	s.balances[b.Header.Miner] += BlockReward

	return nil
}

func applyTrx(trx SignedTrx, s *State) error {
	ok, err := trx.IsAuthentic()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("wrong transaction. Sender '%s' is forged", trx.From.String())
	}
	if len(trx.From) == 0 {
		return NewInvalidTransaction("From")
	}
	if len(trx.To) == 0 {
		return NewInvalidTransaction("To")
	}
	if trx.Value == 0 {
		return NewInvalidTransaction("Value")
	}

	if trx.Value > s.balances[trx.From] {
		return new(ErrInsufficientBalance)
	}

	s.balances[trx.From] -= trx.Value
	s.balances[trx.To] += trx.Value

	return nil
}

func applyTRXs(trxs []SignedTrx, s *State) error {
	sort.Slice(trxs, func(i, j int) bool {
		return trxs[i].Time < trxs[j].Time
	})

	for _, trx := range trxs {
		err := applyTrx(trx, s)
		if err != nil {
			return err
		}
	}

	return nil
}
