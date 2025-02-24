package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"tbb/v2/database"
)

func Run(dataDir string) error {
	s, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		return err
	}
	defer s.Close()

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			Hash     database.Hash             `json:"block_hash"`
			Balances map[database.Account]uint `json:"balances"`
		}{
			Hash:     s.LatestHash(),
			Balances: s.Balances,
		}

		payloadJson, err := json.Marshal(payload)
		if err != nil {
			errRes := struct {
				Error string `json:"error"`
			}{err.Error()}

			jsonErrRes, err := json.Marshal(errRes)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write(jsonErrRes)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("unable to read request body: %v\n", err)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("unable to unmarshal request body: %v\n", err)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("error adding transaction: %v", err)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			return
		}

		hash, err := s.Persist()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("error persisting transaction: %v", err)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			return
		}

		res := struct {
			Hash database.Hash `json:"block_hash"`
		}{hash}
		resJson, err := json.Marshal(res)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("unable to marshal response body: %s", err)))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
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
