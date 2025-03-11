package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const BlockReward = 100

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
		Parent Hash    `json:"parent"`
		Height uint64  `json:"height"`
		Nonce  uint32  `json:"nonce"`
		Time   uint64  `json:"time"`
		Miner  Account `json:"miner"`
	}

	Hash [32]byte
)

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.Hex()), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) IsEmpty() bool {
	return bytes.Equal(h[:], []byte(new(Hash)[:]))
}

func (h Hash) IsValid() bool {
	return fmt.Sprintf("%x", h[0]) == "0" &&
		fmt.Sprintf("%x", h[1]) == "0" &&
		fmt.Sprintf("%x", h[2]) == "0" &&
		fmt.Sprintf("%x", h[3]) != "0"
}

func NewBlock(parent Hash, height uint64, nonce uint32, time uint64, miner Account, trxs []Trx) Block {
	return Block{BlockHeader{parent, height, nonce, time, miner}, trxs}
}

func (b Block) Hash() (Hash, error) {
	blockJSON, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(blockJSON), nil
}
