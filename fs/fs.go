package fs

import "github.com/spf13/afero"

const (
	Dir     = "database"
	GenFile = "genesis.json"
	TrxFile = "block.db"
)

var AppFS *afero.Afero

func init() {
	AppFS = &afero.Afero{Fs: afero.NewOsFs()}
}
