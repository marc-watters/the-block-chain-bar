package database

type ErrInsufficientBalance struct {
	From  Account
	To    Account
	Value uint64
}

func (e ErrInsufficientBalance) Error() string {
	return "insufficient balance"
}
