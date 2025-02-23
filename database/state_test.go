package database_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"tbb/v2/database"
	"testing"
	"time"

	"github.com/spf13/afero"
)

var appFs *afero.Afero

func init() {
	database.AppFs = &afero.Afero{Fs: afero.NewMemMapFs()}
	appFs = database.AppFs
}

func init() {
	genData := []byte(`{"balances": {"A": 0, "B": 1}}`)
	if err := appFs.WriteFile(genF, genData, os.ModeAppend); err != nil {
		fmt.Printf("error writing to genesis file: %v\n", err)
		os.Exit(1)
	}

	if err := appFs.WriteFile(txF, []byte{}, os.ModeAppend); err != nil {
		fmt.Printf("error writing to transaction file: %v\n", err)
		os.Exit(1)
	}
}

var (
	dir  = database.Dir
	genF = filepath.Join(dir, database.GenF)
	txF  = filepath.Join(dir, database.TxF)
)

var (
	a = database.NewAccount("A")
	b = database.NewAccount("B")
)

func TestNewStateFromDisk(t *testing.T) {
	t.Run("assert accounts and balances from new state", func(t *testing.T) {
		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		assertAccount(t, s, a, 0)
		assertAccount(t, s, b, 1)
	})

	t.Run("assert error insufficent balance", func(t *testing.T) {
		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		block := database.NewBlock(
			database.Hash{},
			uint64(time.Now().Unix()),
			[]database.Tx{
				database.NewTx(a, b, 1, ""),
			},
		)

		err = s.AddBlock(block)
		if err == nil {
			t.Fatal("assert insufficent balance failed, expected an error")
		}

		var insufficientBalance *database.ErrInsufficientBalance
		if !errors.As(err, &insufficientBalance) {
			t.Errorf("assert insufficient balance failed, unexpected error: %T", err)
		}
	})

	t.Run("assert state add transaction", func(t *testing.T) {
		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		block := database.NewBlock(
			database.Hash{},
			uint64(time.Now().Unix()),
			[]database.Tx{
				database.NewTx(b, a, 1, ""),
			},
		)

		if err := s.AddBlock(block); err != nil {
			t.Fatalf("error adding block: %v", err)
		}

		assertAccount(t, s, a, 1)
		assertAccount(t, s, b, 0)
	})

	t.Run("assert state persist transactions", func(t *testing.T) {
		s, err := database.NewStateFromDisk()
		if err != nil {
			t.Fatalf("error loading state: %v", err)
		}

		parentBlock := database.NewBlock(
			database.Hash{},
			uint64(time.Now().Unix()),
			[]database.Tx{
				database.NewTx(b, a, 1, ""),
			},
		)

		if err := s.AddBlock(parentBlock); err != nil {
			t.Fatalf("error adding block: %v", err)
		}

		parentHash, err := s.Persist()
		if err != nil {
			t.Fatalf("error persisting block A: %v", err)
		}

		childBlock := database.NewBlock(
			parentHash,
			uint64(time.Now().Unix()),
			[]database.Tx{
				database.NewTx(a, b, 1, ""),
			},
		)

		err = s.AddBlock(childBlock)
		if err != nil {
			t.Fatalf("error adding block B: %v", err)
		}

		childHash, err := s.Persist()
		if err != nil {
			t.Fatalf("error persisting blockB: %v", err)
		}

		got, err := appFs.ReadFile(txF)
		if err != nil {
			t.Fatalf("error reading block.db file: %v", err)
		}

		bfsParent, err := json.Marshal(database.BlockFS{parentHash, parentBlock})
		if err != nil {
			t.Fatalf("error marshaling parent blockFS: %v", err)
		}

		bfsChild, err := json.Marshal(database.BlockFS{childHash, childBlock})
		if err != nil {
			t.Fatalf("error marshaling child blockFS: %v", err)
		}

		want := slices.Concat(append(bfsParent, '\n'), append(bfsChild, '\n'))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("assert persisted transactions failed:\ngot:\n%d\nwant:\n%d", got, want)
		}
	})
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
