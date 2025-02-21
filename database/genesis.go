package database

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}
