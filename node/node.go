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

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

const httpPort = 8080

func Run(dataDir string) error {
	s, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		return err
	}
	defer s.Close()

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, s)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, s)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
}

func listBalancesHandler(w http.ResponseWriter, _ *http.Request, s *database.State) {
	writeRes(w, BalanceRes{s.LatestHash(), s.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, s *database.State) {
	req := TxAddReq{}
	err := readReq(r, &req)
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

	writeRes(w, TxAddRes{hash})
}

func readReq(r *http.Request, reqBody any) error {
	reqBodyJson, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, &reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body: %s", err.Error())
	}

	return nil
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
