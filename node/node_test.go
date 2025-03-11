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
	"strings"
	"testing"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

type mockState struct {
	latestBlock     db.Block
	latestBlockHash db.Hash
	balances        map[db.Account]uint64
	trxMempool      []db.Trx
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

func (ms *mockState) DataDir() string {
	return "./"
}

func (ms *mockState) LatestBlock() db.Block {
	return ms.latestBlock
}

func (ms *mockState) LatestBlockHash() db.Hash {
	return ms.latestBlockHash
}

func (ms *mockState) NextBlockHeight() uint64 {
	panic("not implemented")
}

func (ms *mockState) Balances() map[db.Account]uint64 {
	return ms.balances
}

func (ms *mockState) AddBlock(b db.Block) (db.Hash, error) {
	hash, err := b.Hash()
	if err != nil {
		return db.Hash{}, nil
	}

	ms.latestBlockHash = hash
	ms.latestBlock = b

	return hash, nil
}

func (ms *mockState) AddBlocks(blocks []db.Block) error {
	// TODO: Implement mockState.AddBlocks()
	panic("not implemented")
}

func TestNode(t *testing.T) {
	t.Run("get balances", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/balances/list", nil)
		res := httptest.NewRecorder()

		n := &Node{state: newMockState()}
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
			n.state.LatestBlockHash(),
			n.state.Balances(),
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetBalances() = %+v, want %+v", got, want)
		}
	})

	t.Run("post transaction", func(t *testing.T) {
		body := strings.NewReader(`{"from":"A","to":"B","value":1,"data":""}`)
		req, _ := http.NewRequest(http.MethodPost, "/tx/add", body)
		res := httptest.NewRecorder()

		n := &Node{state: newMockState()}
		n.PostTrx(res, req)

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("PostTrx() read error: %v", err)
		}

		var got TrxPostRes
		if err := json.Unmarshal(resBody, &got); err != nil {
			t.Fatalf("PostTrx() unmarshal error: %v", err)
		}

		want := TrxPostRes{n.state.LatestBlockHash()}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("PostTrx() = %+v, want %+v", got, want)
		}
	})
}
