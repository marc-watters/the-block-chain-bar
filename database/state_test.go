package database_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

var (
	appFS *afero.Afero

	dir     = "./"
	genFile = filepath.Join(fs.Dir, fs.GenFile)
	trxFile = filepath.Join(fs.Dir, fs.TrxFile)
)

func init() {
	fs.AppFS = &afero.Afero{Fs: afero.NewMemMapFs()}
	appFS = fs.AppFS
}

func TestNewStateFromDisk(t *testing.T) {
	t.Run("assert new state accounts and balances", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances":{"a": 1,"b": 0}}`),
			/* transaction */ []byte(``),
		)

		got := s.Balances
		want := map[db.Account]uint64{"a": 1, "b": 0}

		assertBalance(t, got, want)
	})

	t.Run("assert error insufficient balance", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances":{"a": 0,"b": 0}}`),
			/* transaction */ []byte(``),
		)

		var got error
		var want *db.ErrInsufficientBalance

		got = s.AddTrx(db.Trx{
			From:  "a",
			To:    "b",
			Value: 1,
		})

		if got == nil {
			t.Fatalf("State.Add() error = %v, wanted %v", got, want)
		}
		if !errors.As(got, &want) {
			t.Errorf("State.Add() error type = %T, wanted %T", got, want)
		}
	})

	t.Run("assert add single transaction", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances": {"a": 1, "b": 0}}`),
			/* transaction */ []byte(``),
		)

		err := s.AddTrx(db.Trx{
			From:  db.NewAccount("a"),
			To:    db.NewAccount("b"),
			Value: 1,
		})
		checkError(t, "State.Add", err)

		_, err = s.Persist()
		checkError(t, "State.Persist", err)

		got := s.Balances
		want := map[db.Account]uint64{"a": 0, "b": 1}

		assertBalance(t, got, want)
	})

	t.Run("assert add multiple transactions", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances": {"a": 1, "b": 0}}`),
			/* transaction */ []byte(``),
		)

		a := db.NewAccount("a")
		b := db.NewAccount("b")
		trxs := []db.Trx{
			{From: a, To: b, Value: 1},
			{From: b, To: a, Value: 1},
		}

		for i := range trxs {
			err := s.AddTrx(trxs[i])
			checkError(t, "State.Add", err)
		}

		_, err := s.Persist()
		checkError(t, "State.Persist", err)

		got := s.Balances
		want := map[db.Account]uint64{"a": 1, "b": 0}

		assertBalance(t, got, want)
	})

	t.Run("assert add invalid transaction", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances":{"A": 1,"B": 0}}`),
			/* transaction */ []byte(``),
		)

		tests := []struct {
			name string
			trx  db.Trx
			want error
		}{
			{
				"zero value field: 'From'",
				db.Trx{From: "", To: "A", Value: 1},
				db.NewInvalidTransaction("From"),
			},
			{
				"zero value field: 'To'",
				db.Trx{From: "A", To: "", Value: 1},
				db.NewInvalidTransaction("To"),
			},
			{
				"zero value field: 'value'",
				db.Trx{From: "A", To: "B", Value: 0},
				db.NewInvalidTransaction("Value"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := s.AddTrx(tt.trx)
				if got == nil {
					t.Fatalf("State.Add() error = %v, wanted %s", got, tt.want)
				}
				if got != tt.want {
					t.Errorf("State.Add() error = %v, wanted %s", got, tt.want)
				}
			})
		}
	})
}

func composeState(t testing.TB, genData, trxData []byte) *db.State {
	t.Helper()

	if err := appFS.WriteFile(genFile, genData, os.ModeAppend); err != nil {
		t.Fatalf("error writing to genesis file: %v", err)
	}

	if err := appFS.WriteFile(trxFile, trxData, os.ModeAppend); err != nil {
		t.Fatalf("error writing to transaction file: %v", err)
	}

	s, err := db.NewStateFromDisk(dir)
	if err != nil {
		t.Fatalf("NewStateFromDisk() error = %v", err)
	}

	return s
}

func assertBalance(t testing.TB, got, want map[db.Account]uint64) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("State.Balances = %+v, want %+v", got, want)
	}
}

func checkError(t testing.TB, caller string, err error) {
	if err != nil {
		t.Fatalf("%s() error = %v", caller, err)
	}
}
