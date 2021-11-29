package light

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	"github.com/prysmaticlabs/prysm/io/file"
	"github.com/prysmaticlabs/prysm/testing/require"
	"testing"

	tmplog "log"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func hashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func hex0x(b []byte) string {
	return "0x" + hex.EncodeToString(b[:])
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

func Test_NextSyncCommitteeBranch(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := ztype.FromSszBytes(sszbytes)

	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(23)
	leaf := state.GetLeaf(gIndex)
	nextSyncCommitteeBranch := state.GetBranch(gIndex)

	tmplog.Println(root)
	tmplog.Println(gIndex)
	tmplog.Println(leaf)
	tmplog.Println(nextSyncCommitteeBranch)
}

func Test_FinalityCheckpointBranch(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := ztype.FromSszBytes(sszbytes)

	//{"finalized_checkpoint", CheckpointType}, // 20

	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(20) // finalized_checkpoint
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	require.Equal(t, ztype.Verify(root, gIndex, leaf, branch), true)
	tmplog.Println(root)
	tmplog.Println(gIndex)
	tmplog.Println(leaf)
	tmplog.Println(branch)
}

func Test_FinalityHeaderBranch(t *testing.T) {
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/state/aafc599c3a8980bc4daa867ee2e8a52a88219761.ssz"
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := ztype.FromSszBytes(sszbytes)

	root := state.HashTreeRoot()
	gIndex := state.GetGIndex(20, 1) // finalized_checkpoint (20), root (1)
	leaf := state.GetLeaf(gIndex)
	branch := state.GetBranch(gIndex)

	require.Equal(t, ztype.Verify(root, gIndex, leaf, branch), true)
	require.Equal(t, leaf, state.State.FinalizedCheckpoint().Root) // The merkle root of the leaf is the leaf itself.

	tmplog.Println("leaf                ", leaf)
	tmplog.Println("finalized checkpoint", hex0x(state.State.FinalizedCheckpoint().Root))
}
