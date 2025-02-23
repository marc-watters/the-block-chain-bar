package database_test

import (
	"fmt"
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

func TestNewStateFromDisk(t *testing.T) {}
