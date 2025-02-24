package database

import "path/filepath"

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, Dir)
}
