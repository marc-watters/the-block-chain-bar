package database

import (
	"os"
	"path/filepath"
)

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
