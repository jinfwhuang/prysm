package light

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"

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
	log "github.com/sirupsen/logrus"
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
	root := update.AttestedHeader.StateRoot
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
	root := update.AttestedHeader.StateRoot
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

	root1 := bytesutil.ToBytes32(update.AttestedHeader.StateRoot)
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

func (s *Service) learnState(ctx context.Context, root []byte) {
	tmplog.Println("-------learning-----------")
	tmplog.Println("root", base64.StdEncoding.EncodeToString(root))
	blk, err := s.getBlock(ctx, root)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	//header, err := blk.Header()
	//if err != nil {
	//	tmplog.Println(err)
	//	log.Error(err)
	//	return
	//}
	blkRoot, err := blk.Block().HashTreeRoot()
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("block root", base64.StdEncoding.EncodeToString(blkRoot[:]))

	st, err := s.getState(ctx, root)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	stRoot, err := st.HashTreeRoot(context.Background())
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("state root", base64.StdEncoding.EncodeToString(stRoot[:]))
	tmplog.Println("checkpoint root", base64.StdEncoding.EncodeToString(st.FinalizedCheckpoint().Root))
	tmplog.Println("state checkpoint", st.FinalizedCheckpoint())

	fRoot := st.FinalizedCheckpoint().Root

	fBlk, err := s.getBlock(ctx, fRoot)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("---")
	tmplog.Println("f root", base64.StdEncoding.EncodeToString(fRoot))
	tmplog.Println("f block", fBlk)

	fSt, err := s.getState(ctx, fRoot)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	fStRoot, err := fSt.HashTreeRoot(context.Background())
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("f state root", base64.StdEncoding.EncodeToString(fStRoot[:]))
	tmplog.Println("----------------")
}

func (s *Service) _GetChainHeadAndState(ctx context.Context) (block.SignedBeaconBlock, state.BeaconState, error) {
	root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
	if err != nil {
		return nil, nil, err
	}
	blk, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(root))
	if err != nil {
		return nil, nil, err
	}
	if blk == nil || blk.IsNil() {
		return nil, nil, fmt.Errorf("head is nil: blockRoot=%s", base64.StdEncoding.EncodeToString(root))
	}

	st, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(root))
	if err != nil {
		return nil, nil, err
	}
	if st == nil || st.IsNil() {
		tmplog.Println(s.cfg.Database.HasState(ctx, bytesutil.ToBytes32(root)))
		return nil, nil, fmt.Errorf("head state is nil: blockRoot=%s", base64.StdEncoding.EncodeToString(root))
	}

	return blk, st, nil
}

func (s *Service) parseFinalizedEvent(ctx context.Context, eventData interface{}) (block.SignedBeaconBlock, state.BeaconState, error) {
	finalizedCheckpoint, ok := eventData.(*v1.EventFinalizedCheckpoint)
	if !ok {
		return nil, nil, fmt.Errorf("expected finalized checkpoint event")
	}
	blk, err := s.getBlock(ctx, finalizedCheckpoint.Block)
	if err != nil {
		return nil, nil, err
	}
	st, err := s.getState(ctx, finalizedCheckpoint.Block)
	return blk, st, nil
}
