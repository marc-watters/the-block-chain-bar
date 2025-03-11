package node

import (
	"context"
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

	acc := db.NewAccount("andrej")
	pn := NewPeerNode(DefaultIP, peerNodePort, true, acc, true)
	n := New(s, DefaultIP, DefaultHTTPort, acc, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if err := n.Run(ctx); err.Error() != "http: Server closed" {
		cancel()
		t.Errorf("expected node to shutdown after 5 seconds")
	}
	cancel()
}

func TestNode_Mining(t *testing.T) {
	dataDir := getTestDataDirPath(t)
	if err := fs.RemoveDir(dataDir); err != nil {
		t.Fatal(err)
	}

	andrejAcc := db.NewAccount("andrej")
	babayagaAcc := db.NewAccount("babayaga")

	s, err := db.NewStateFromDisk(dataDir)
	if err != nil {
		t.Fatalf("error getting new state from disk: %v", err)
	}

	n := New(s, DefaultIP, 8085, babayagaAcc, PeerNode{})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)

	trx := db.Trx{From: "andrej", To: "babayaga", Value: 1, Time: 1579451695, Data: ""}
	trx2 := db.NewTrx("andrej", "babayaga", 2, "")

	trx2hash, err := trx2.Hash()
	if err != nil {
		t.Fatal("error hashing transaction 2: %v", err)
		cancel()
	}

	validSyncedBlock := db.NewBlock(
		db.Hash{},
		uint64(1),
		uint32(1275873026),
		uint64(1580415832),
		db.NewAccount("andrej"),
		[]db.Trx{trx},
	)

	go func() {
		time.Sleep((mininingIntervalSeconds - 2) * time.Second)

		myself := NewPeerNode(DefaultIP, 8085, false, db.NewAccount(""), true)
		if err := n.AddPendingTrx(trx, myself); err != nil {
			t.Fatalf("error adding pending transaction 1: %v", err)
			cancel()
		}

		if err := n.AddPendingTrx(trx2, myself); err != nil {
			t.Fatalf("error adding pending transaction 2: %v", err)
			cancel()
		}
	}()

	go func() {
		time.Sleep((mininingIntervalSeconds + 2) * time.Second)
		if !n.isMining {
			t.Fatal("should be mining")
			cancel()
		}

		if _, err := n.state.AddBlock(validSyncedBlock); err != nil {
			t.Fatal("error adding valid synced block: %v", err)
			cancel()
		}

		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(2 * time.Second)
		if n.isMining {
			t.Fatal("new received block should have cancelled mining")
			cancel()
		}

		_, onlyTrx2IsPending := n.pendingTRXs[trx2hash.Hex()]

		if len(n.pendingTRXs) != 1 && !onlyTrx2IsPending {
			t.Fatal("new received block should have cancelled mining of already mined transaction")
			cancel()
		}

		time.Sleep((mininingIntervalSeconds + 2) * time.Second)

		if !n.isMining {
			t.Fatal("should be mining again, transaction 1 not included in synced block")
			cancel()
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Height == 2 {
					closeNode()
					return
				}
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

		expectedEndAndrejBalance := startingAndrejBalance - trx.Value - trx2.Value + db.BlockReward
		expectedEndBabayagaBalance := startingBabayagaBalance + trx.Value + trx2.Value + db.BlockReward

		if endAndrejBalance != expectedEndAndrejBalance {
			t.Errorf("Andrej expected end balance is %d not %d", expectedEndAndrejBalance, endAndrejBalance)
		}

		if endBabayagaBalance != expectedEndBabayagaBalance {
			t.Errorf("Babayaga expected end balance is %d not %d", expectedEndBabayagaBalance, endBabayagaBalance)
		}

		t.Logf("Starting Andrej balance: %d", startingAndrejBalance)
		t.Logf("Starting Babayaga balance: %d", startingBabayagaBalance)
		t.Logf("Ending Andrej balance: %d", endAndrejBalance)
		t.Logf("Ending Babayaga balance: %d", endBabayagaBalance)
	}()

	if err := n.Run(ctx); err != nil {
		t.Fatalf("error running node: %v", err)
	}

	if n.state.LatestBlock().Header.Height != 2 {
		t.Fatal("was supposed to mine 2 pending transactions into 2 valid blocks under 30m")
	}

	if len(n.pendingTRXs) != 0 {
		t.Fatal("no pending transactions should be left to mine")
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
