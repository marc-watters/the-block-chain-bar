package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/fs"
	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
)

// The password for testing keystore files:
//
//	./test_andrej--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57
//	./test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8
//
// Pre-generated for testing purposes using wallet_test.go.
//
// It's necessary to have pre-existing accounts before a new node
// with fresh new, empty keystore is initialized and booted in order
// to configure the accounts balances in genesis.json
//
// I.e: A quick solution to a chicken-egg problem.
const (
	testKsAndrejAccount   = "0x3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
	testKsBabaYagaAccount = "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
	testKsAndrejFile      = "test_andrej--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
	testKsBabaYagaFile    = "test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
	testKsAccountsPwd     = "security123"

	peerNodePort = 8081
)

func TestNode_Run(t *testing.T) {
	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("error getting test data directory path: %v", err)
	}

	if err := fs.RemoveDir(dataDir); err != nil {
		t.Fatalf("error cleansing data directory: %v", err)
	}

	s, err := db.NewStateFromDisk(dataDir)
	if err != nil {
		t.Fatalf("error creating new state from disk: %v", err)
	}

	n := New(s, DefaultIP, DefaultHTTPort, db.NewAccount(wallet.AndrejAccount), PeerNode{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := n.Run(ctx); err != nil {
		t.Errorf("expected node to shutdown after 5 seconds")
	}
}

func TestNode_Mining(t *testing.T) {
	dataDir, andrej, babayaga, err := setupTestNodeDir()
	if err != nil {
		t.Fatalf("error setting up test node directory: %v", err)
	}

	pn := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		babayaga,
		true,
	)

	s, err := db.NewStateFromDisk(dataDir)
	if err != nil {
		t.Fatalf("error getting new state from disk: %v", err)
	}

	n := New(s, pn.IP, pn.Port, andrej, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	errC := make(chan error)
	ticker := time.NewTicker(10 * time.Second)

	// Schedule a new TX in 3 seconds from now, in a separate thread
	// because the n.Run() few lines below is a blocking call
	go func() {
		time.Sleep((mininingIntervalSeconds / 3) * time.Second)
		trx := db.NewTrx(andrej, babayaga, 1, "")
		signedTrx, err := wallet.SignTrxWithKeystoreAccount(trx, andrej, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			errC <- fmt.Errorf("error creating new signed transaction: %v", err)
		}
		errC <- n.AddPendingTrx(signedTrx, pn)
	}()
	// Schedule a new TX in 12 seconds from now simulating
	// that it came in - while the first TX is being mined
	go func() {
		time.Sleep((mininingIntervalSeconds + 2) * time.Second)
		trx := db.NewTrx(andrej, babayaga, 2, "")
		signedTrx, err := wallet.SignTrxWithKeystoreAccount(trx, andrej, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			errC <- fmt.Errorf("error creating new signed transaction: %v", err)
		}
		errC <- n.AddPendingTrx(signedTrx, pn)
	}()
	// Periodically check if we mined the 2 blocks
	go func() {
		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Height == 1 {
					cancel()
					errC <- nil
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-errC:
				if err != nil {
					fmt.Println("error received:", err)
					cancel()
				}
			case <-ctx.Done():
				break
			}
		}
	}()

	if err := n.Run(ctx); err != nil {
		t.Fatalf("error running node: %v", err)
	}

	if n.state.LatestBlock().Header.Height != 1 {
		t.Fatal("2 pending transactions not mined into 2 blocks under 30m")
	}
}

// The test logic summary:
//   - BabaYaga runs the node
//   - BabaYaga tries to mine 2 TXs
//   - The mining gets interrupted because a new block from Andrej gets synced
//   - Andrej will get the block reward for this synced block
//   - The synced block contains 1 of the TXs BabaYaga tried to mine
//   - BabaYaga tries to mine 1 TX left
//   - BabaYaga succeeds and gets her block reward
func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	andrej := db.NewAccount(testKsAndrejAccount)
	babayaga := db.NewAccount(testKsBabaYagaAccount)

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("error getting test data directory path: %v", err)
	}

	genesisBalances := make(map[common.Address]uint64)
	genesisBalances[andrej] = 1000000
	genesis := db.Genesis{Balances: genesisBalances}
	genesisJSON, err := json.Marshal(genesis)
	if err != nil {
		t.Fatalf("error marshalling genesis: %v", err)
	}

	if err := fs.InitDataDirIfNotExists(dataDir, genesisJSON); err != nil {
		t.Fatalf("error initializing data directory: %v", err)
	}
	defer func() {
		err := fs.RemoveDir(dataDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error removing data directory: %v", err)
		}
	}()

	if err := copyKeystoreFilesIntoTestDataDirPath(dataDir); err != nil {
		t.Fatalf("error copying keystore files: %v", err)
	}

	s, err := db.NewStateFromDisk(dataDir)
	if err != nil {
		t.Fatalf("error getting new state from disk: %v", err)
	}

	pn := NewPeerNode(
		DefaultIP,
		peerNodePort,
		false,
		db.NewAccount(""),
		true,
	)

	n := New(s, pn.IP, pn.Port, babayaga, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	trx1 := db.NewTrx(andrej, babayaga, 1, "")
	trx2 := db.NewTrx(andrej, babayaga, 2, "")

	signedTrx1, err := wallet.SignTrxWithKeystoreAccount(trx1, andrej, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Fatalf("error signing transaction 1 for andrej: %v", err)
	}

	signedTrx2, err := wallet.SignTrxWithKeystoreAccount(trx1, andrej, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Fatalf("error signing transaction 2 for babyaga: %v", err)
	}

	trx2hash, err := trx2.Hash()
	if err != nil {
		t.Fatalf("error hashing transaction 2: %v", err)
		cancel()
	}

	validPreMinedPb := NewPendingBlock(db.Hash{}, 0, andrej, []db.SignedTrx{signedTrx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatalf("error mining block: %v", err)
		cancel()
	}

	errC := make(chan error)
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		time.Sleep((mininingIntervalSeconds - 2) * time.Second)
		errC <- n.AddPendingTrx(signedTrx1, pn)
		errC <- n.AddPendingTrx(signedTrx2, pn)
	}()

	go func() {
		time.Sleep((mininingIntervalSeconds + 2) * time.Second)

		if !n.isMining {
			errC <- fmt.Errorf("should be mining")
		}

		if _, err := n.state.AddBlock(validSyncedBlock); err != nil {
			errC <- fmt.Errorf("error adding block: %v", err)
		}

		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(2 * time.Second)
		if n.isMining {
			errC <- fmt.Errorf("synced block should have cancelled mining")
		}

		_, onlyTrx2IsPending := n.pendingTRXs[trx2hash.Hex()]

		if len(n.pendingTRXs) != 1 && !onlyTrx2IsPending {
			errC <- fmt.Errorf("synced block should have cancelled mining of already mined block")
		}

		time.Sleep((mininingIntervalSeconds + 2) * time.Second)
		if !n.isMining {
			errC <- fmt.Errorf("should be mining the 1x transaction not included in synced block")
		}
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Height == 1 {
					cancel()
					return
				}
			case err := <-errC:
				if err != nil {
					fmt.Println("go routine error:", err)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)

		startingAndrejBalance := n.state.Balances()[andrej]
		startingBabayagaBalance := n.state.Balances()[babayaga]

		<-ctx.Done()

		endAndrejBalance := n.state.Balances()[andrej]
		endBabayagaBalance := n.state.Balances()[babayaga]

		expectedEndAndrejBalance := startingAndrejBalance - trx1.Value - trx2.Value + db.BlockReward
		expectedEndBabayagaBalance := startingBabayagaBalance + trx1.Value + trx2.Value + db.BlockReward

		if endAndrejBalance != expectedEndAndrejBalance {
			errC <- fmt.Errorf("Andrej expected balance is %d not %d", expectedEndAndrejBalance, endAndrejBalance)
			cancel()
			return

		}

		if endBabayagaBalance != expectedEndBabayagaBalance {
			errC <- fmt.Errorf("Babayaga expected balance is %d not %d", expectedEndBabayagaBalance, endBabayagaBalance)
			cancel()
			return
		}

		t.Logf("Starting Andrej balance: %d", startingAndrejBalance)
		t.Logf("Starting Babayaga balance: %d", startingBabayagaBalance)
		t.Logf("Ending Andrej balance: %d", endAndrejBalance)
		t.Logf("Ending Babayaga balance: %d", endBabayagaBalance)
	}()

	if err := n.Run(ctx); err != nil {
		t.Fatalf("error during node run: %v", err)
		cancel()
	}
}

func getTestDataDirPath() (string, error) {
	return fs.AppFS.TempDir(os.TempDir(), "tbb_test")
}

// Copy the pre-generated, commited keystore files from this folder into the new testDataDirPath()
//
// Afterwards the test datadir path will look like:
//
//	"/tmp/tbb_test945924586/keystore/test_andrej--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
//	"/tmp/tbb_test945924586/keystore/test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	andrejSrcKs, err := fs.AppFS.Open(testKsAndrejFile)
	if err != nil {
		return err
	}
	defer andrejSrcKs.Close()

	ksDir := filepath.Join(wallet.GetKeystoreDirPath(dataDir))

	err = os.Mkdir(ksDir, 0777)
	if err != nil {
		return err
	}

	andrejDstKs, err := os.Create(filepath.Join(ksDir, testKsAndrejFile))
	if err != nil {
		return err
	}
	defer andrejDstKs.Close()

	_, err = io.Copy(andrejDstKs, andrejSrcKs)
	if err != nil {
		return err
	}

	babayagaSrcKs, err := os.Open(testKsBabaYagaFile)
	if err != nil {
		return err
	}
	defer babayagaSrcKs.Close()

	babayagaDstKs, err := os.Create(filepath.Join(ksDir, testKsBabaYagaFile))
	if err != nil {
		return err
	}
	defer babayagaDstKs.Close()

	_, err = io.Copy(babayagaDstKs, babayagaSrcKs)
	if err != nil {
		return err
	}

	return nil
}

// setupTestNodeDir creates a default testing node directory with 2 keystore accounts
// Remember to remove the dir once test finishes: defer fs.RemoveDir(dataDir)
func setupTestNodeDir() (dataDir string, andrej, babaYaga common.Address, err error) {
	babaYaga = db.NewAccount(testKsBabaYagaAccount)
	andrej = db.NewAccount(testKsAndrejAccount)

	dataDir, err = getTestDataDirPath()
	if err != nil {
		return "", common.Address{}, common.Address{}, fmt.Errorf("getting test data directory failed: %w", err)
	}
	fmt.Println("datadir", dataDir)

	genesisBalances := make(map[common.Address]uint64)
	genesisBalances[andrej] = 1000000
	genesis := db.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		return "", common.Address{}, common.Address{}, fmt.Errorf("marshalling genesis failed: %w", err)
	}

	err = fs.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil {
		return "", common.Address{}, common.Address{}, fmt.Errorf("initializing data directory failed: %w", err)
	}

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		return "", common.Address{}, common.Address{}, fmt.Errorf("copying key store failed: %w", err)
	}

	return dataDir, andrej, babaYaga, nil
}
