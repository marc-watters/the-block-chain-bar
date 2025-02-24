package database

import (
	"os"
	"path/filepath"
)

const genesisJson = `
{
  "genesis_time": "2019-03-18T00:00:00.000000000Z",
  "chain_id": "the-blockchain-bar-ledger",
  "balances": {
    "andrej": 1000000
  }
}`
func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, Dir)
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), GenF)
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), TxF)
}

func fileExist(filePath string) bool {
	_, err := AppFs.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func writeGenesisToDisk(path string) error {
	return AppFs.WriteFile(path, []byte(genesisJson), 0644)
}
