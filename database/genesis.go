package database

import (
	"encoding/json"
)

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := AppFs.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var g genesis
	err = json.Unmarshal(content, &g)
	if err != nil {
		return genesis{}, err
	}

	return g, nil
}
