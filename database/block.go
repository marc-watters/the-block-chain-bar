package database

import "encoding/hex"

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

type Block struct {
	Header  BlockHeader // metadata (parent block hash + time)
	Payload []Tx        // new transactions only (payload)
}

func NewBlock(parent Hash, time uint64, txs []Tx) Block {
	return Block{BlockHeader{parent, time}, txs}
}
