package node

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type mockState struct {
	latestBlockHash db.Hash
	balances        map[db.Account]uint64
}

func newMockState() *mockState {
	ms := &mockState{
		balances: map[db.Account]uint64{
			db.NewAccount("A"): 1,
			db.NewAccount("B"): 0,
		},
	}

	if _, err := rand.Read(ms.latestBlockHash[:]); err != nil {
		fmt.Printf("error generating random hash: %v\n", err)
		os.Exit(1)
	}

	return ms
}

func (ms *mockState) LatestBlockHash() db.Hash {
	return ms.latestBlockHash
}

func (ms *mockState) Balances() map[db.Account]uint64 {
	return ms.balances
}

func TestNode(t *testing.T) {
	t.Run("get balances", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/balances/list", nil)
		res := httptest.NewRecorder()

		n := &Node{newMockState()}
		n.GetBalances(res, req)

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("GetBalances() read error: %v", err)
		}

		var got BalanceRes
		if err := json.Unmarshal(resBody, &got); err != nil {
			t.Fatalf("GetBalances() unmarhsal error: %v", err)
		}

		want := BalanceRes{
			n.LatestBlockHash(),
			n.Balances(),
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetBalances() = %+v, want %+v", got, want)
		}
	})
}
