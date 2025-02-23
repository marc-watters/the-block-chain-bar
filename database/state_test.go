package database_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"tbb/v2/database"
	"testing"

	"github.com/spf13/afero"
)

var appFs *afero.Afero

func init() {
	database.AppFs = &afero.Afero{Fs: afero.NewMemMapFs()}
	appFs = database.AppFs
}

var (
	dir  = database.Dir
	genF = filepath.Join(dir, database.GenF)
	txF  = filepath.Join(dir, database.TxF)
)

func TestNewStateFromDisk(t *testing.T) {
	t.Run("assert accounts and balances from new state", func(t *testing.T) {
		composeStateFiles(t,
			/* genesis     */ []byte(`{"balances": {"A": 1, "B": 0}}`),
			/* transaction */ []byte(`{"from": "A", "to": "B", "value": 1}`),
		)

		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		a := database.NewAccount("A")
		b := database.NewAccount("B")

		assertAccount(t, s, a, 0)
		assertAccount(t, s, b, 1)
	})

	t.Run("assert error insufficent balance", func(t *testing.T) {
		composeStateFiles(t,
			/* genesis     */ []byte(`{"balances": {"A": 1, "B": 0}}`),
			/* transaction */ []byte(`{"from": "A", "to": "B", "value": 2}`),
		)
		_, err := database.NewStateFromDisk()
		if err == nil {
			t.Errorf("assert insufficient balance failed, expected an error")
		}

		var insufficientBalance database.ErrInsufficientBalance
		if !errors.As(err, &insufficientBalance) {
			t.Errorf("assert insufficient balance failed, unexpected error: %v", err)
		}
	})

	t.Run("assert state add transaction", func(t *testing.T) {
		composeStateFiles(t,
			/* genesis     */ []byte(`{"balances": {"A": 1, "B": 0}}`),
			/* transaction */ []byte(``),
		)

		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		a := database.NewAccount("A")
		b := database.NewAccount("B")

		if err := s.Add(database.Tx{From: a, To: b, Value: 1}); err != nil {
			t.Fatalf("error adding transaction: %v", err)
		}

		assertAccount(t, s, a, 0)
		assertAccount(t, s, b, 1)
	})

	t.Run("assert state persist transactions", func(t *testing.T) {
		composeStateFiles(t,
			/* genesis     */ []byte(`{"balances": {"A": 1, "B": 0}}`),
			/* transaction */ []byte(``),
		)

		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		a := database.NewAccount("A")
		b := database.NewAccount("B")

		cases := []struct {
			from  database.Account
			to    database.Account
			value uint
			data  string
		}{
			{a, b, 1, ""},
			{b, a, 1, ""},
		}

		txs := make([]database.Tx, 0)
		for _, c := range cases {
			tx := database.NewTx(c.from, c.to, c.value, c.data)

			if err := s.Add(tx); err != nil {
				t.Fatalf("error adding transaction: %v", err)
			}
			if _, err := s.Persist(); err != nil {
				t.Fatalf("error persisting transaction: %v", err)
			}

			txs = append(txs, tx)

			assertPersistedTxs(t, txs)
		}
	})
}

func composeStateFiles(t testing.TB, genData, txData []byte) {
	t.Helper()

	if err := appFs.WriteFile(genF, genData, os.ModeAppend); err != nil {
		t.Fatalf("error writing to genesis file: %v", err)
	}

	if err := appFs.WriteFile(txF, txData, os.ModeAppend); err != nil {
		t.Fatalf("error writing to transaction file: %v", err)
	}
}

func assertAccount(t testing.TB, s *database.State, a database.Account, bal uint) {
	t.Helper()

	val, ok := s.Balances[a]
	if !ok {
		t.Errorf("assert account failed: could not find account %q", a)
	}

	if val != bal {
		t.Errorf("assert balance failed: wrong balance for %q: got %d, want %d", a, val, bal)
	}
}

func assertPersistedTxs(t testing.TB, txs []database.Tx) {
	got, err := appFs.ReadFile(txF)
	if err != nil {
		t.Fatalf("error reading transaction file: %v", err)
	}

	want := func() []byte {
		b := bytes.NewBuffer([]byte{})
		for _, tx := range txs {
			b.Write([]byte(fmt.Sprint(tx.AsJson(), "\n")))
		}

		return b.Bytes()
	}

	if !reflect.DeepEqual(got, want()) {
		t.Errorf("assert sequential persisted transaction failed:\n\tgot: \t%v\n\twant:\t%v", got, want())
	}
}
