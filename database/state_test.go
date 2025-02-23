package database_test

import (
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
		genData := []byte(`{"balances": {"A": 0, "B": 1}}`)
		if err := appFs.WriteFile(genF, genData, os.ModeAppend); err != nil {
			t.Fatalf("error writing to genesis file: %v", err)
		}

		if err := appFs.WriteFile(txF, []byte{}, os.ModeAppend); err != nil {
			t.Fatalf("error writing to transaction file: %v", err)
		}

		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		a := database.NewAccount("A")
		b := database.NewAccount("B")

		aBal, ok := s.Balances[a]
		if !ok {
			t.Errorf("assert account failed: could not find account %q", a)
		}

		if aBal != 0 {
			t.Errorf("assert balance failed: wrong balance for %q: got %d, want %d", a, aBal, 0)
		}

		bBal, ok := s.Balances[b]
		if !ok {
			t.Errorf("assert account failed: could not find account %q", b)
		}

		if bBal != 1 {
			t.Errorf("assert balance failed: wrong balance for %q: got %d, want %d", b, bBal, 1)
		}
	})
