package database

import "encoding/json"

type genesis struct {
	Balances map[Account]uint64 `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := AppFS.ReadFile(path)
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
