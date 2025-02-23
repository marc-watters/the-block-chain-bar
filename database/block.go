package database

import "encoding/hex"

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}
