package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"tbb/v2/database"
)

type BalanceRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type ErrRes struct {
	Error string `json:"error"`
}

func Run(dataDir string) error {
	s, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		return err
	}
	defer s.Close()

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		payload := BalanceRes{s.LatestHash(), s.Balances}
		payloadJson, err := json.Marshal(payload)
		if err != nil {
			writeErrRes(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(payloadJson)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrRes(w, err)
			return
		}
		defer r.Body.Close()

		req := struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Value uint   `json:"value"`
			Data  string `json:"data"`
		}{}
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			writeErrRes(w, err)
			return
		}

		tx := database.NewTx(
			database.NewAccount(req.From),
			database.NewAccount(req.To),
			req.Value,
			req.Data,
		)
		err = s.AddTx(tx)
		if err != nil {
			writeErrRes(w, err)
			return
		}

		hash, err := s.Persist()
		if err != nil {
			writeErrRes(w, err)
			return
		}

		res := struct {
			Hash database.Hash `json:"block_hash"`
		}{hash}
		resJson, err := json.Marshal(res)
		if err != nil {
			writeErrRes(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resJson)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	})

	return http.ListenAndServe(":8080", nil)
}

func writeErrRes(w http.ResponseWriter, err error) {
	jsonErrRes, err := json.Marshal(ErrRes{err.Error()})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(jsonErrRes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
