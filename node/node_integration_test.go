package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

const peerNodePort = 8081

func TestNode_Run(t *testing.T) {
	dataDir := getTestDataDirPath(t)
	if err := fs.RemoveDir(dataDir); err != nil {
		t.Fatalf("error cleansing data directory: %v", err)
	}

	s, err := db.NewStateFromDisk(dataDir)
	if err != nil {
		t.Fatalf("error creating new state from disk: %v", err)
	}

	n := New(s, DefaultIP, DefaultHTTPort, db.NewAccount("andrej"), PeerNode{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := n.Run(ctx); err != nil {
		t.Errorf("expected node to shutdown after 5 seconds")
	}
}

func TestNode_Mining(t *testing.T) {
	dataDir := getTestDataDirPath(t)
	if err := fs.RemoveDir(dataDir); err != nil {
		t.Fatal(err)
	}

	andrejAcc := db.NewAccount("andrej")

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

	n := New(s, pn.IP, pn.Port, andrejAcc, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	errC := make(chan error)
	ticker := time.NewTicker(10 * time.Second)

	// Schedule a new TX in 3 seconds from now, in a separate thread
	// because the n.Run() few lines below is a blocking call
	go func() {
		time.Sleep((mininingIntervalSeconds / 3) * time.Second)
		trx := db.NewTrx("andrej", "babayaga", 1, "")
		errC <- n.AddPendingTrx(trx, pn)
	}()
	// Schedule a new TX in 12 seconds from now simulating
	// that it came in - while the first TX is being mined
	go func() {
		time.Sleep((mininingIntervalSeconds + 2) * time.Second)
		trx := db.NewTrx("andrej", "babayaga", 2, "")
		if !n.isMining {
			errC <- fmt.Errorf("should be mining")
			cancel()
		} else {
			errC <- n.AddPendingTrx(trx, pn)
		}
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

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	dataDir := getTestDataDirPath(t)
	if err := fs.RemoveDir(dataDir); err != nil {
		t.Fatal(err)
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

	andrejAcc := db.NewAccount("andrej")
	babayagaAcc := db.NewAccount("babayaga")

	n := New(s, pn.IP, pn.Port, babayagaAcc, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	trx1 := db.NewTrx("andrej", "babayaga", 1, "")
	trx2 := db.NewTrx("andrej", "babayaga", 2, "")

	trx2hash, err := trx2.Hash()
	if err != nil {
		t.Fatalf("error hashing transaction 2: %v", err)
		cancel()
	}

	validPreMinedPb := NewPendingBlock(db.Hash{}, 0, andrejAcc, []db.Trx{trx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatalf("error mining block: %v", err)
		cancel()
	}

	errC := make(chan error)
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		time.Sleep((mininingIntervalSeconds - 2) * time.Second)
		errC <- n.AddPendingTrx(trx1, pn)
		errC <- n.AddPendingTrx(trx2, pn)
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

		startingAndrejBalance := n.state.Balances()[andrejAcc]
		startingBabayagaBalance := n.state.Balances()[babayagaAcc]

		<-ctx.Done()

		endAndrejBalance := n.state.Balances()[andrejAcc]
		endBabayagaBalance := n.state.Balances()[babayagaAcc]

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

func getTestDataDirPath(t testing.TB) string {
	t.Helper()

	dir, err := fs.AppFS.TempDir("./", ".tbb_test")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	return dir
}
