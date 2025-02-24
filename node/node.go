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
		writeRes(w, BalanceRes{s.LatestHash(), s.Balances})
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
		writeRes(w, res)
	})

	return http.ListenAndServe(":8080", nil)
}

func writeRes(w http.ResponseWriter, payload any) {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(payloadJson); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
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
