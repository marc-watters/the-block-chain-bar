package database

import "strings"

type Account string

func NewAccount(value string) Account {
	return Account(strings.ToLower(value))
}
