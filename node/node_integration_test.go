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

	pn := NewPeerNode(DefaultIP, peerNodePort, true, true)
	n := New(s, DefaultIP, DefaultHTTPort, pn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if err := n.Run(ctx); err.Error() != "http: Server closed" {
		cancel()
		t.Errorf("expected node to shutdown after 5 seconds")
	}
	cancel()
}

func getTestDataDirPath(t testing.TB) string {
	t.Helper()

	dir, err := fs.AppFS.TempDir("./", ".tbb_test")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}
	return dir
}
