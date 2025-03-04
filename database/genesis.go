package database

type genesis struct {
	Balances map[string]uint64 `json:"balances"`
}
