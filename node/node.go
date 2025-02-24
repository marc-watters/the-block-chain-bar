package node

import (
	"encoding/json"
	"fmt"
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

	return http.ListenAndServe(":8080", nil)
}
