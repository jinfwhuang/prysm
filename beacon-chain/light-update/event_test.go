package light

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	"github.com/prysmaticlabs/prysm/io/file"
	"github.com/prysmaticlabs/prysm/proto/eth/v2"
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

func Hex0x(b []byte) string {
	return "0x" + hex.EncodeToString(b[:])
}

func HexStr(b []byte) string {
	return hex.EncodeToString(b[:])
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

	gIndex := state.GetGIndex(23)
	root := state.HashTreeRoot()
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

func Test_FinalityHeaderBranch2(t *testing.T) {
	// State
	_filename := "08eb3e781e76cf041406c71c5c48567874b8f1068c31f5f8324701e5278a128d.ssz"
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/beacon-state/" + _filename
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	state := ztype.FromSszBytes(sszbytes)
	//stateRoot := state.HashTreeRoot()

	// Finality Block
	_filename = "147e2adcac6cf5424d6d36251b247b2633483b8d51e917742a88e527c9d83778.ssz"
	filename = "/Users/jin/code/repos/prysm/tmp/ssz/keep/block/" + _filename
	sszbytes, err = file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	//block := wrapper.altairBeaconBlock{}
	finalityBlock := eth.BeaconBlockAltair{}
	err = finalityBlock.UnmarshalSSZ(sszbytes)
	require.NoError(t, err)
	finalityBlockRoot, err := finalityBlock.HashTreeRoot()
	require.NoError(t, err)

	// Check
	require.Equal(t, HexStr(state.State.FinalizedCheckpoint().Root), HexStr(finalityBlockRoot[:]))

	// Branch verification
	_root := state.HashTreeRoot()
	_gIndex := state.GetGIndex(20, 1) // finalized_checkpoint (20), root (1)
	_leaf := state.GetLeaf(_gIndex)
	_branch := state.GetBranch(_gIndex)
	tmplog.Println("verification        ", ztype.Verify(_root, _gIndex, _leaf, _branch))
	tmplog.Println("leaf                ", _leaf)
	tmplog.Println("finalized checkpoint", hex0x(state.State.FinalizedCheckpoint().Root))
	tmplog.Println("finalityBlock root  ", hex0x(finalityBlockRoot[:]))
}

func Test_NumOfSetBits(t *testing.T) {
	//_filename := "08eb3e781e76cf041406c71c5c48567874b8f1068c31f5f8324701e5278a128d.ssz"
	//filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/beacon-state/" + _filename
	//sszbytes, err := file.ReadFileAsBytes(filename)
	//require.NoError(t, err)
	//state := ztype.FromSszBytes(sszbytes)

	_filename := "147e2adcac6cf5424d6d36251b247b2633483b8d51e917742a88e527c9d83778.ssz"
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/keep/block/" + _filename
	sszbytes, err := file.ReadFileAsBytes(filename)
	require.NoError(t, err)
	//block := wrapper.altairBeaconBlock{}
	finalityBlock := eth.BeaconBlockAltair{}
	err = finalityBlock.UnmarshalSSZ(sszbytes)
	require.NoError(t, err)

	syncAgg := finalityBlock.Body.SyncAggregate

	bitField := syncAgg.SyncCommitteeBits

	numOfSetBits(bitField)
}
