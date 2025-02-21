package database_test

import (
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
	gd := []byte(`{"balances": {"A": 1, "B": 0}}`)
	if err := appFs.WriteFile(genF, gd, 0600); err != nil {
		t.Fatalf("error writing to %q: %v", genF, err)
	}

	txd := []byte(`{"from": "A", "to": "B", "value": 1}`)
	if err := appFs.WriteFile(txF, txd, 0600); err != nil {
		t.Fatalf("error writing to %q: %v", txF, err)
	}

	s, err := database.NewStateFromDisk()
	if err != nil {
		t.Fatalf("error loading state: %v", err)
	}

	a, ok := s.Balances["A"]
	if !ok {
		t.Errorf("assert state account failed: could not find account 'A'")
	}

	b, ok := s.Balances["B"]
	if !ok {
		t.Errorf("assert state account failed: could not find account 'B'")
	}

	if a != 0 {
		t.Errorf("assert state balance failed: wrong balance for account 'A': %d", a)
	}

	if b != 1 {
		t.Errorf("assert state balance failed: wrong balance for account 'B': %d", a)
	}
}
