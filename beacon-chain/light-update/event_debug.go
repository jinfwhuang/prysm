package light

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"

	//vv1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	statev2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	tmplog "log"
	"os"
	//"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	//"github.com/prysmaticlabs/prysm/encoding/ssz"
	//"github.com/prysmaticlabs/prysm/network/forks"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"

	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype/utils"
)

func hexStr(b []byte) string {
	return hex.EncodeToString(b[:])
}

func writetofile(filename string, root [32]byte, bytearray []byte) {
	err := os.WriteFile(filename, bytearray, 0666)
	if err != nil {
		panic(err)
	}

	contentHash := utils.HashBytes(bytearray)

	tmplog.Println("----------------------------")
	tmplog.Println("wrote a ssz file:", filename)
	tmplog.Println("root            :", hex.EncodeToString(root[:]))
	tmplog.Println("content hash    :", contentHash)
	tmplog.Println("----------------------------")
}

type Bbb interface {
	MarshalSSZ() ([]byte, error)
}

func save(s Bbb) {

}

func saveSsz(block block.BeaconBlock, ztypeState *ztype.ZtypBeaconStateAltair, finalityBlock block.BeaconBlock) {
	// block
	bytearrayBlock, err := block.MarshalSSZ()
	PanicErr(err)
	rootBlock, err := block.HashTreeRoot()
	PanicErr(err)
	filenameBlock := "./tmp/ssz/keep/block/" + hexStr(rootBlock[:]) + ".ssz"
	writetofile(filenameBlock, rootBlock, bytearrayBlock)

	// finality block
	bytearrayFinality, err := finalityBlock.MarshalSSZ()
	PanicErr(err)
	rootFinality, err := block.HashTreeRoot()
	PanicErr(err)
	filenameFinality := "./tmp/ssz/keep/block/" + hexStr(rootFinality[:]) + ".ssz"
	writetofile(filenameFinality, rootFinality, bytearrayFinality)

	// beacon state
	bytearray := ztypeState.SszSerialize()
	stateRoot := ztypeState.HashTreeRoot()
	filename := "./tmp/ssz/keep/beacon-state/" + hexStr(stateRoot[:]) + ".ssz"
	writetofile(filename, stateRoot, bytearray)
}

func verifyMerkleFinalityHeader(update *ethpb.LightClientUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 20, 1) // next_sync_committee  23
	root := update.Header.StateRoot
	leaf, err := update.FinalityHeader.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin
	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.FinalityBranch
	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

func verifyMerkleNextSyncComm(update *ethpb.LightClientUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 23) // next_sync_committee  23
	root := update.Header.StateRoot
	leaf, err := update.NextSyncCommittee.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin

	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.NextSyncCommitteeBranch
	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

func testVerification(update *ethpb.LightClientUpdate, ztypeState ztype.ZtypBeaconStateAltair) {
	// Testing a verification
	ver := verifyMerkleFinalityHeader(update)
	tmplog.Println("ver", ver)

	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 20, 1) // next_sync_committee  23

	root1 := bytesutil.ToBytes32(update.FinalityHeader.StateRoot)
	leaf1, _ := update.FinalityHeader.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin
	tmplog.Println("using update", ztype.Verify(root1, gIndex, leaf1, update.NextSyncCommitteeBranch))

	root2 := ztypeState.HashTreeRoot()
	leaf2, _ := update.FinalityHeader.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin
	tmplog.Println("second try", ztype.Verify(root2, gIndex, leaf2, update.FinalityBranch))

	tmplog.Println("root1", base64.StdEncoding.EncodeToString(root1[:]))
	tmplog.Println("root2", base64.StdEncoding.EncodeToString(root2[:]))
	tmplog.Println("leaf1", base64.StdEncoding.EncodeToString(leaf1[:]))
	tmplog.Println("leaf2", base64.StdEncoding.EncodeToString(leaf2[:]))
	tmplog.Println("leaf2", base64.StdEncoding.EncodeToString(ztypeState.State.FinalizedCheckpoint().Root))

}

//func saveSsz(ctx context.Context, state ) {
//	// Use beaconstatev2.BeaconState
//	bytearray, err := state.MarshalSSZ()
//	if err != nil {
//		panic(err)
//	}
//	stateRoot, err := state.HashTreeRoot(ctx)
//	if err != nil {
//		panic(err)
//	}
//	tmplog.Println("statev2")
//	writetofile(bytearray, stateRoot)
//
//	// Use ethpb.BeaconStateAltair
//	ethpbState := state.(*statev2.BeaconState).GetState()
//	bytearray, err = ethpbState.MarshalSSZ()
//	if err != nil {
//		panic(err)
//	}
//	ethpbStateRoot, err := ethpbState.HashTreeRoot()
//	if err != nil {
//		panic(err)
//	}
//	tmplog.Println("ethpb BeaconState")
//	writetofile(bytearray, ethpbStateRoot)
//}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func verify_serder_same_root(ctx context.Context, state1 state.BeaconState) {
	root1, err := state1.HashTreeRoot(ctx)
	stateStructv1 := state1.(*statev2.BeaconState)
	PanicErr(err)

	bytearray, err := state1.MarshalSSZ()
	state2 := mkBeaconStateAltair(bytearray) // From ser-der process
	stateStructv2 := state2.(*statev2.BeaconState)
	root2, err := state2.HashTreeRoot(ctx)

	tmplog.Println(hex.EncodeToString(root1[:]))
	tmplog.Println(hex.EncodeToString(root2[:]))

	tmplog.Println(stateStructv1.GetState().LatestBlockHeader.BodyRoot)
	tmplog.Println(stateStructv2.GetState().LatestBlockHeader.BodyRoot)

	tmplog.Println(stateStructv1.GetState().LatestBlockHeader.StateRoot)
	tmplog.Println(stateStructv2.GetState().LatestBlockHeader.StateRoot)
}

func mkBeaconStateAltair(bytearray []byte) state.BeaconState {
	ethpbState := &ethpb.BeaconStateAltair{}
	err := ethpbState.UnmarshalSSZ(bytearray)
	PanicErr(err)

	state, err := statev2.InitializeFromProto(ethpbState)
	PanicErr(err)

	return state
}

// TODO: hack
func hex0x(b []byte) string {
	return "0x" + hex.EncodeToString(b[:])
}

func testVerifyNexSynComm(update *ethpb.LightClientUpdate, ztypeState ztype.ZtypBeaconStateAltair) {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 23) // next_sync_committee  23

	// Testing a verification
	ver := verifyMerkleNextSyncComm(update)
	tmplog.Println("ver", ver)

	root1 := bytesutil.ToBytes32(update.Header.StateRoot)
	leaf1, _ := update.NextSyncCommittee.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin
	tmplog.Println("using update", ztype.Verify(root1, gIndex, leaf1, update.NextSyncCommitteeBranch))

	root2 := ztypeState.HashTreeRoot()
	leaf2 := ztypeState.GetLeaf(gIndex)
	tmplog.Println("second try", ztype.Verify(root2, gIndex, leaf2, update.NextSyncCommitteeBranch))

	tmplog.Println("root1", base64.StdEncoding.EncodeToString(root1[:]))
	tmplog.Println("root2", base64.StdEncoding.EncodeToString(root2[:]))
	tmplog.Println("leaf1", base64.StdEncoding.EncodeToString(leaf1[:]))
	tmplog.Println("leaf2", base64.StdEncoding.EncodeToString(leaf2[:]))

}
