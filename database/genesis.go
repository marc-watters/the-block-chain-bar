package database

import (
	"encoding/json"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

type genesis struct {
	Balances map[Account]uint64 `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := fs.AppFS.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}
