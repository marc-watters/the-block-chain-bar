package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type (
	BlockFS struct {
		Key   Hash  `json:"hash"`
		Value Block `json:"block"`
	}
	Block struct {
		Header BlockHeader `json:"header"`
		TRXs   []Trx       `json:"payload"`
	}
	BlockHeader struct {
		Parent Hash   `json:"parent"`
		Time   uint64 `json:"time"`
	}

	Hash [32]byte
)

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

func NewBlock(parent Hash, time uint64, trxs []Trx) Block {
	return Block{BlockHeader{parent, time}, trxs}
}

func (b Block) Hash() (Hash, error) {
	blockJSON, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(blockJSON), nil
}
