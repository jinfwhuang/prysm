package ztype

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/testing/require"
	"math/rand"

	///Users/jin/code/repos/prysm/beacon-chain/state/v2

	//"bytes"
	//"github.com/protolambda/ztyp/codec"
	//"github.com/prysmaticlabs/go-bitfield"
	//"github.com/prysmaticlabs/prysm/beacon-chain/state"
	//"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/io/file"
	//"github.com/prysmaticlabs/prysm/testing/util"
	tmplog "log"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func hashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func TestAA(t *testing.T) {
	tmplog.Println("fff")
}

//func Test_StateConversion(t *testing.T) {
//	tmplog.Println("fff")
//	//st, _ := util.NewBeaconState()
//	//dec := codec.NewDecodingReader(bytes.NewReader(data), uint64(len(data)))
//	//convertState()
//}

func Test_HashTreeRoot(t *testing.T) {
	//st, _ := util.NewBeaconState()
	//dec := codec.NewDecodingReader(bytes.NewReader(data), uint64(len(data)))
	//convertState()
	//filename := "/Users/jin/code/repos/prysm/tmp/ssz/statev1/24691.ssz" // The fastssz hash is: 62113.ssz
	ctx := context.Background()
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// Calculate root through the ztyp library
	root1 := state.HashTreeRoot()

	// Calculate root through the algorithm in: beacon-chain/state/v2/state_trie.go
	root2, err := state.State.HashTreeRoot(ctx)
	require.NoError(t, err)

	// Validation
	require.Equal(t, hex.EncodeToString(root1[:]), hex.EncodeToString(root2[:]))

	tmplog.Println(hex.EncodeToString(root1[:]))
	tmplog.Println(hex.EncodeToString(root2[:]))
	tmplog.Println("----")
	tmplog.Println(root1[:])
	tmplog.Println(root2[:])
	tmplog.Println("----")

}

func Test_SszSerialize(t *testing.T) {
	//ctx := context.Background()
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// Through ztyp library
	sszbytes1 := state.SszSerialize()

	// Through fastssz library
	sszbytes2, err := state.State.MarshalSSZ()
	require.NoError(t, err)

	n := rand.Intn(len(sszbytes1)) // Get some random bytes
	tmplog.Println(n)
	tmplog.Println(sszbytes1[n : n+17])
	tmplog.Println(sszbytes2[n : n+17])

	tmplog.Println(hashBytes(sszbytes1))
	tmplog.Println(hashBytes(sszbytes2))
}

func Test_Verify(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// {"next_sync_committee", SyncCommitteeType}, // 23
	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(23)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	okay := Verify(root, gIndex, leaf, branch)

	tmplog.Println(root)
	tmplog.Println(gIndex)
	tmplog.Println(leaf)
	tmplog.Println(branch)
	tmplog.Println(okay)
}

func Test_Proof(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := FromSszBytes(sszbytes)

	// {"next_sync_committee", SyncCommitteeType}, // 23
	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(23)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	tmplog.Println(root)
	tmplog.Println(gIndex)
	tmplog.Println(leaf)
	tmplog.Println(branch)

}
