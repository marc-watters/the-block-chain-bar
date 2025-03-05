package fs

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	Dir     = "database"
	GenFile = "genesis.json"
	TrxFile = "block.db"
)

var AppFS *afero.Afero

func init() {
	AppFS = &afero.Afero{Fs: afero.NewOsFs()}
}

var genesisJSON = `
{
  "genesis_time": "2019-03-18T00:00:00.000000000Z",
  "chain_id": "the-blockchain-bar-ledger",
  "balances": {
    "andrej": 1000000
  }
}`

func GetDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, Dir)
}

func GetGenesisJSONFilePath(dataDir string) string {
	return filepath.Join(GetDatabaseDirPath(dataDir), GenFile)
}

func GetBlocksDBFilePath(dataDir string) string {
	return filepath.Join(GetDatabaseDirPath(dataDir), TrxFile)
}

func FileExist(path string) bool {
	_, err := AppFS.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func DirExists(path string) (bool, error) {
	_, err := AppFS.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
