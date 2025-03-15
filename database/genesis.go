package database

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

type Genesis struct {
	Balances map[common.Address]uint64 `json:"balances"`
}

func loadGenesis(path string) (Genesis, error) {
	content, err := fs.AppFS.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}
