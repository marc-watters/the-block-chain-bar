package database_test

import (
	"fmt"
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
	t.Run("assert new state from disk", func(t *testing.T) {
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

		// test one transaction
		tx := database.NewTx(a, b, 1, "")
		if err := s.Add(tx); err != nil {
			t.Fatalf("error adding transaction: %v", err)
		}

		assertAccount(t, s, a, 0)
		assertAccount(t, s, b, 1)

		if err := s.Persist(); err != nil {
			t.Fatalf("error persisting transaction: %v", err)
		}

		got, err := appFs.ReadFile(txF)
		if err != nil {
			t.Fatalf("error reading transaction file: %v", err)
		}

		want := []byte("{\"from\":\"A\",\"to\":\"B\",\"value\":1,\"data\":\"\"}\n")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("assert persisted transaction failed:\n\tgot: \t%q\n\twant:\t%q", got, want)
		}

		// test sequential transactions
		tx = database.NewTx(b, a, 1, "")
		if err := s.Add(tx); err != nil {
			t.Fatalf("error adding transaction: %v", err)
		}

		assertAccount(t, s, a, 1)
		assertAccount(t, s, b, 0)

		if err := s.Persist(); err != nil {
			t.Fatalf("error persisting transaction: %v", err)
		}

		got, err = appFs.ReadFile(txF)
		if err != nil {
			t.Fatalf("error reading transaction file: %v", err)
		}

		want = []byte(fmt.Sprint(
			"{\"from\":\"A\",\"to\":\"B\",\"value\":1,\"data\":\"\"}\n",
			"{\"from\":\"B\",\"to\":\"A\",\"value\":1,\"data\":\"\"}\n",
		))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("assert sequential persisted transaction failed:\n\tgot: \t%q\n\twant:\t%q", got, want)
		}
	})
}

func composeStateFiles(t testing.TB, genData, txData []byte) {
	t.Helper()

	if err := appFs.WriteFile(genF, genData, 0600); err != nil {
		t.Fatalf("error writing to genesis file: %v", err)
	}

	if err := appFs.WriteFile(txF, txData, 0600); err != nil {
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
