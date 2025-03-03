package database_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

var appFS *afero.Afero

func init() {
	db.AppFS = &afero.Afero{Fs: afero.NewMemMapFs()}
	appFS = db.AppFS
}

func TestNewStateFromDisk(t *testing.T) {
	t.Run("assert new state accounts and balances", func(t *testing.T) {
		genData := []byte(`{"balances":{"a": 1,"b": 0}}`)
		genFile := filepath.Join("database", "genesis.json")
		err := appFS.WriteFile(genFile, genData, 0o400)
		if err != nil {
			t.Fatalf("error writing genesis file: %v", err)
		}

		trxData := []byte(``)
		trxFile := filepath.Join("database", "trx.db")
		err = appFS.WriteFile(trxFile, trxData, 0o400)
		if err != nil {
			t.Fatalf("error writing transaction file: %v", err)
		}

		s, err := db.NewStateFromDisk()
		if err != nil {
			t.Fatalf("NewStateFromDisk() error = %v", err)
		}

		got := s.Balances
		want := map[db.Account]uint64{
			"a": 1,
			"b": 0,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("State.Balances = %v, want %v", got, want)
		}
	})
}
