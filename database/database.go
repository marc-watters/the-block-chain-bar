package database

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

func GetBlocksAfter(blockHash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(fs.GetBlocksDBFilePath(dataDir), os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)

	var shouldStartCollecting bool
	if reflect.DeepEqual(blockHash, Hash{}) {
		shouldStartCollecting = true
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, err
		}

		var blockFs BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blockFs)
		if err != nil {
			return nil, err
		}

		if shouldStartCollecting {
			blocks = append(blocks, blockFs.Value)
			continue
		}

		if blockHash == blockFs.Key {
			shouldStartCollecting = true
		}
	}

	return blocks, nil
}
