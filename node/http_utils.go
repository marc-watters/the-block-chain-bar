package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type (
	BalanceRes struct {
		Hash     db.Hash               `json:"block_hash"`
		Balances map[db.Account]uint64 `json:"balances"`
	}
	ErrRes struct {
		Error string `json:"error"`
	}
	TrxPostReq struct {
		From  db.Account `json:"from"`
		To    db.Account `json:"to"`
		Value uint64     `json:"value"`
		Data  string     `json:"data"`
	}
	TrxPostRes struct {
		Hash db.Hash `json:"block_hash"`
	}
	StatusRes struct {
		Hash       db.Hash             `json:"block_hash"`
		Height     uint64              `json:"block_height"`
		KnownPeers map[string]PeerNode `json:"peers_known"`
	}
	SyncRes struct {
		Blocks []db.Block `json:"blocks"`
	}
	AddPeerRes struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
)

func writeRes(w http.ResponseWriter, data any) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		writeErr(w, err)
		return
	}

	if _, err := w.Write(dataJSON); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func writeErr(w http.ResponseWriter, err error) {
	errJSON, err := json.Marshal(ErrRes{err.Error()})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	http.Error(w, string(errJSON), http.StatusInternalServerError)
}

func readReq(r *http.Request, reqBody any) error {
	reqBodyJSON, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %v", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(reqBodyJSON, &reqBody); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %v", err)
	}

	return nil
}

func readRes(r *http.Response, reqBody any) error {
	reqBodyJSON, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body. %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJSON, &reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body. %s", err.Error())
	}

	return nil
}
