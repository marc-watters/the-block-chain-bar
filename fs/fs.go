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

var genesisJSON = `
{
  "genesis_time": "2019-03-18T00:00:00.000000000Z",
  "chain_id": "the-blockchain-bar-ledger",
  "balances": {
    "andrej": 1000000
  }
}`
