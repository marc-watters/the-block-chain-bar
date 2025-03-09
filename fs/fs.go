package fs

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

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

func InitDataDirIfNotExists(dataDir string) error {
	if FileExist(GetGenesisJSONFilePath(dataDir)) {
		return nil
	}

	if err := AppFS.MkdirAll(GetDatabaseDirPath(dataDir), os.ModePerm); err != nil {
		return err
	}

	if err := WriteGenesisToDisk(GetGenesisJSONFilePath(dataDir)); err != nil {
		return err
	}

	if err := WriteEmptyBlocksDBToDisk(GetBlocksDBFilePath(dataDir)); err != nil {
		return err
	}

	return nil
}

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

func WriteEmptyBlocksDBToDisk(path string) error {
	return AppFS.WriteFile(path, []byte(``), os.ModePerm)
}

func WriteGenesisToDisk(path string) error {
	return AppFS.WriteFile(path, []byte(genesisJSON), 0o644)
}

func RemoveDir(path string) error {
	return AppFS.RemoveAll(path)
}

func ExpandPath(p string) string {
	if i := strings.Index(p, ":"); i > 0 {
		return p
	}
	if i := strings.Index(p, "@"); i > 0 {
		return p
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := homeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
