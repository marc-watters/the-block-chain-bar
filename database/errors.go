package database

import "fmt"

type (
	ErrInsufficientBalance struct {
		From  Account
		To    Account
		Value uint64
	}
	ErrInvalidTransaction struct {
		field string
	}
)

func NewInvalidTransaction(field string) ErrInvalidTransaction {
	return ErrInvalidTransaction{field}
}

func (e ErrInsufficientBalance) Error() string {
	return "insufficient balance"
}

func (e ErrInvalidTransaction) Error() string {
	return fmt.Sprintf("invalid value for field: '%s'", e.field)
}
