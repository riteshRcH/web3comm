package coreunix

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	files "github.com/ipfs/go-ipfs-files"
	pi "github.com/ipfs/go-ipfs-posinfo"
	config "github.com/ipfs/go-ipfs/config"
	dag "github.com/ipfs/go-merkledag"
)

const testPeerID = "QmTFauExutTsy4XP6JbMFcw2Wa9645HJt2bTqL6qYDCKfe"

func testAddWPosInfo(t *testing.T, rawLeaves bool) {
	r := &repo.Mock{
		C: config.Config{
			Identity: config.Identity{
				PeerID: testPeerID, // required by offline node
			},
		},
		D: syncds.MutexWrap(datastore.NewMapDatastore()),
	}
	node, err := core.NewNode(context.Background(), &core.BuildCfg{Repo: r})
	if err != nil {
		t.Fatal(err)
	}

	bs := &testBlockstore{GCBlockstore: node.Blockstore, expectedPath: filepath.Join(os.TempDir(), "foo.txt"), t: t}
	bserv := blockservice.New(bs, node.Exchange)
	dserv := dag.NewDAGService(bserv)
	adder, err := NewAdder(context.Background(), node.Pinning, bs, dserv)
	if err != nil {
		t.Fatal(err)
	}
	out := make(chan interface{})
	adder.Out = out
	adder.Progress = true
	adder.RawLeaves = rawLeaves
	adder.NoCopy = true

	data := make([]byte, 5*1024*1024)
	rand.New(rand.NewSource(2)).Read(data) // Rand.Read never returns an error
	fileData := ioutil.NopCloser(bytes.NewBuffer(data))
	fileInfo := dummyFileInfo{"foo.txt", int64(len(data)), time.Now()}
	file, _ := files.NewReaderPathFile(filepath.Join(os.TempDir(), "foo.txt"), fileData, &fileInfo)

	go func() {
		defer close(adder.Out)
		_, err = adder.AddAllAndPin(context.Background(), file)
		if err != nil {
			t.Error(err)
		}
	}()
	for range out {
	}

	exp := 0
	nonOffZero := 0
	if rawLeaves {
		exp = 1
		nonOffZero = 19
	}
	if bs.countAtOffsetZero != exp {
		t.Fatalf("expected %d blocks with an offset at zero (one root and one leaf), got %d", exp, bs.countAtOffsetZero)
	}
	if bs.countAtOffsetNonZero != nonOffZero {
		// note: the exact number will depend on the size and the sharding algo. used
		t.Fatalf("expected %d blocks with an offset > 0, got %d", nonOffZero, bs.countAtOffsetNonZero)
	}
}

func TestAddWPosInfo(t *testing.T) {
	testAddWPosInfo(t, false)
}

func TestAddWPosInfoAndRawLeafs(t *testing.T) {
	testAddWPosInfo(t, true)
}

type testBlockstore struct {
	blockstore.GCBlockstore
	expectedPath         string
	t                    *testing.T
	countAtOffsetZero    int
	countAtOffsetNonZero int
}

func (bs *testBlockstore) Put(ctx context.Context, block blocks.Block) error {
	bs.CheckForPosInfo(block)
	return bs.GCBlockstore.Put(ctx, block)
}

func (bs *testBlockstore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, blk := range blocks {
		bs.CheckForPosInfo(blk)
	}
	return bs.GCBlockstore.PutMany(ctx, blocks)
}

func (bs *testBlockstore) CheckForPosInfo(block blocks.Block) {
	fsn, ok := block.(*pi.FilestoreNode)
	if ok {
		posInfo := fsn.PosInfo
		if posInfo.FullPath != bs.expectedPath {
			bs.t.Fatal("PosInfo does not have the expected path")
		}
		if posInfo.Offset == 0 {
			bs.countAtOffsetZero += 1
		} else {
			bs.countAtOffsetNonZero += 1
		}
	}
}

type dummyFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *dummyFileInfo) Name() string       { return fi.name }
func (fi *dummyFileInfo) Size() int64        { return fi.size }
func (fi *dummyFileInfo) Mode() os.FileMode  { return 0 }
func (fi *dummyFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *dummyFileInfo) IsDir() bool        { return false }
func (fi *dummyFileInfo) Sys() interface{}   { return nil }
