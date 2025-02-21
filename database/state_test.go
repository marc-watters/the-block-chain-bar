package database_test

import (
	"errors"
	"os"
	"path/filepath"
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
