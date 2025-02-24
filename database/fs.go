package database

import (
	"os"
	"os/user"
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

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}

	return ""
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExist(getGenesisJsonFilePath(dataDir)) {
		return nil
	}

	if err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm); err != nil {
		return err
	}

	if err := writeGenesisToDisk(getGenesisJsonFilePath(dataDir)); err != nil {
		return err
	}

	if err := writeEmptyBlocksDbToDisk(getBlocksDbFilePath(dataDir)); err != nil {
		return err
	}

	return nil
}

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

func writeEmptyBlocksDbToDisk(path string) error {
	return AppFs.WriteFile(path, []byte{}, os.ModePerm)
}
