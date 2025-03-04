package database_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

var (
	appFS *afero.Afero

	genFile = filepath.Join(db.Dir, db.GenFile)
	trxFile = filepath.Join(db.Dir, db.TrxFile)
)

func init() {
	db.AppFS = &afero.Afero{Fs: afero.NewMemMapFs()}
	appFS = db.AppFS
}

func TestNewStateFromDisk(t *testing.T) {
	t.Run("assert new state accounts and balances", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances":{"a": 1,"b": 0}}`),
			/* transaction */ []byte(``),
		)

		got := s.Balances
		want := map[db.Account]uint64{
			"a": 1,
			"b": 0,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("State.Balances = %v, want %v", got, want)
		}
	})

	t.Run("assert error insufficient balance", func(t *testing.T) {
		s := composeState(t,
			/* genesis     */ []byte(`{"balances":{"a": 0,"b": 0}}`),
			/* transaction */ []byte(``),
		)

		var got error
		var want *db.ErrInsufficientBalance

		got = s.Add(db.Trx{
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

		err := s.Add(db.Trx{
			From:  db.NewAccount("a"),
			To:    db.NewAccount("b"),
			Value: 1,
		})
		if err != nil {
			t.Fatalf("State.Add() error = %v", err)
		}

		err = s.Persist()
		if err != nil {
			t.Fatalf("State.Persist() error = %v", err)
		}

		got := s.Balances
		want := map[db.Account]uint64{"a": 0, "b": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("State.Balances = %+v, want = %+v", got, want)
		}
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
			err := s.Add(trxs[i])
			if err != nil {
				t.Fatalf("State.Add() error = %v", err)
			}
		}

		err := s.Persist()
		if err != nil {
			t.Fatalf("State.Persist() error = %v", err)
		}

		got := s.Balances
		want := map[db.Account]uint64{"a": 1, "b": 0}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("State.Balances = %+v, want = %+v", got, want)
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

	s, err := db.NewStateFromDisk()
	if err != nil {
		t.Fatalf("NewStateFromDisk() error = %v", err)
	}

	return s
}
