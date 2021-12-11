package light

import "C"
import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	statev1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	statev2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	tmplog "log"
)

func (s *Service) getNonFinalizedSkipSyncUpdate(ctx context.Context, root []byte) (*ethpb.SkipSyncUpdate, error) {
	blk, err := s.getBlock(ctx, root)
	if err != nil {
		return nil, err
	}
	st, err := s.getState(ctx, root)
	if err != nil {
		return nil, err
	}

	attestedHeader, err := blk.Header()
	if err != nil {
		return nil, err
	}

	switch v := st.(type) {
	case *statev1.BeaconState:
		return nil, fmt.Errorf("wrong type: interface conversion: state.BeaconState is *v1.BeaconState, not *v2.BeaconState")
	case *statev2.BeaconState:
		// nothing
	default:
		return nil, fmt.Errorf("unkown type %s", v)
	}
	zState := ztype.FromBeaconState(st.(*statev2.BeaconState))

	syncAgg, err := blk.Block().Body().SyncAggregate()
	if err != nil {
		return nil, err
	}

	currentComm, err := st.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	currentCommBranch := zState.GetBranch(zState.GetGIndex(22))

	nextComm, err := st.NextSyncCommittee()
	if err != nil {
		return nil, err
	}
	nextCommBranch := zState.GetBranch(zState.GetGIndex(23))

	update := &ethpb.SkipSyncUpdate{
		AttestedHeader:             attestedHeader.Header,
		CurrentSyncCommittee:       currentComm,
		CurrentSyncCommitteeBranch: currentCommBranch,
		NextSyncCommittee:          nextComm,
		NextSyncCommitteeBranch:    nextCommBranch,
		FinalityHeader:             &ethpb.BeaconBlockHeader{}, // Empty
		FinalityBranch:             [][]byte{},
		SyncCommitteeBits:          syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature:     syncAgg.SyncCommitteeSignature,
		ForkVersion:                st.Fork().CurrentVersion,
	}

	// TODO: remove
	emptyHeader := &ethpb.BeaconBlockHeader{}
	tmplog.Println(proto.Equal(update.FinalityHeader, emptyHeader))

	return update, nil
}

func (s *Service) getFinalizedSkipSyncUpdate(ctx context.Context, root []byte) (*ethpb.SkipSyncUpdate, error) {
	blk, err := s.getBlock(ctx, root)
	if err != nil {
		return nil, err
	}
	st, err := s.getState(ctx, root)
	if err != nil {
		return nil, err
	}

	attestedHeader, err := blk.Header()
	if err != nil {
		return nil, err
	}
	fCheckpoint := st.FinalizedCheckpoint()
	finalityBlock, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return nil, err
	}
	if finalityBlock == nil {
		return nil, fmt.Errorf("cannot find block with root: %s", base64.StdEncoding.EncodeToString(fCheckpoint.Root))
	}
	signedFinalityHeader, err := finalityBlock.Header()
	finalityHeader := signedFinalityHeader.Header
	if err != nil {
		return nil, err
	}
	zState := ztype.FromBeaconState(st.(*statev2.BeaconState))
	finalityCheckpointRootBranch := zState.GetBranch(zState.GetGIndex(20, 1))

	syncAgg, err := blk.Block().Body().SyncAggregate()
	if err != nil {
		return nil, err
	}

	// Committee information refers to finalize_state
	fSt, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return nil, err
	}
	if fSt == nil {
		return nil, fmt.Errorf("cannot find state with root: %s", base64.StdEncoding.EncodeToString(fCheckpoint.Root))
	}
	zFinalizedState := ztype.FromBeaconState(fSt.(*statev2.BeaconState))
	currentCom, err := fSt.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	currentSyncCommitteeBranch := zFinalizedState.GetBranch(zFinalizedState.GetGIndex(22))

	nextCom, err := fSt.NextSyncCommittee()
	if err != nil {
		return nil, err
	}
	nextSyncCommitteeBranch := zFinalizedState.GetBranch(zFinalizedState.GetGIndex(23))

	skipSyncUpdate := &ethpb.SkipSyncUpdate{
		AttestedHeader:             attestedHeader.Header,
		CurrentSyncCommittee:       currentCom,
		CurrentSyncCommitteeBranch: currentSyncCommitteeBranch,
		NextSyncCommittee:          nextCom,
		NextSyncCommitteeBranch:    nextSyncCommitteeBranch,
		FinalityHeader:             finalityHeader,
		FinalityBranch:             finalityCheckpointRootBranch,
		SyncCommitteeBits:          syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature:     syncAgg.SyncCommitteeSignature,
		ForkVersion:                st.Fork().CurrentVersion,
	}

	return skipSyncUpdate, nil
}

func ToLightClientUpdate(update *ethpb.SkipSyncUpdate) *ethpb.LightClientUpdate {
	return &ethpb.LightClientUpdate{
		AttestedHeader:          update.AttestedHeader,
		NextSyncCommittee:       update.NextSyncCommittee,
		NextSyncCommitteeBranch: update.NextSyncCommitteeBranch,
		FinalityHeader:          update.FinalityHeader,
		FinalityBranch:          update.FinalityBranch,
		SyncCommitteeBits:       update.SyncCommitteeBits,
		SyncCommitteeSignature:  update.SyncCommitteeSignature,
		ForkVersion:             update.ForkVersion,
	}
}

func (s *Service) GetHeadState(ctx context.Context) (state.BeaconState, error) {
	st, err := s.cfg.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, err
	}
	if st == nil || st.IsNil() {
		return nil, fmt.Errorf("cannot get a heade state")
	}
	return st, nil
}

// TODO: improve on how to choose which SkipSyncUpdate to keep
func (s *Service) bestSkipSyncUpdate(ctx context.Context, newUpdate *ethpb.SkipSyncUpdate) *ethpb.SkipSyncUpdate {
	key, err := newUpdate.CurrentSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	update, err := s.GetSkipSyncUpdate(ctx, bytesutil.ToBytes32(key[:]))
	if err != nil || update == nil {
		return newUpdate
	}

	oldParticipation := numOfSetBits(update.SyncCommitteeBits)
	newParticipation := numOfSetBits(newUpdate.SyncCommitteeBits)
	tmplog.Println("oldParticipation", oldParticipation, "newParticipation", newParticipation)

	// Criteria 1: Compare which update has the most participation
	if newParticipation >= oldParticipation {
		return newUpdate
	} else {
		return update
	}
	return newUpdate
}

func (s *Service) getBlock(ctx context.Context, root []byte) (block.SignedBeaconBlock, error) {
	blk, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(root))
	if err != nil {
		return nil, err
	}
	if blk == nil || blk.IsNil() {
		return nil, fmt.Errorf("cannot find block, blockRoot=%s", base64.StdEncoding.EncodeToString(root))
	}

	return blk, nil
}

func (s *Service) getState(ctx context.Context, root []byte) (state.BeaconState, error) {
	st, err := s.cfg.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(root))
	if err != nil {
		return nil, err
	}
	if st == nil || st.IsNil() {
		return nil, fmt.Errorf("state is empty, blockRoot=%s", base64.StdEncoding.EncodeToString(root))
	}

	return st, nil
}

func numOfSetBits(b []byte) int {
	bitStr := fmt.Sprintf("%08b", b)
	// TODO: hack; use bit operation instead
	count := 0
	countZero := 0
	for _, c := range bitStr {
		if c == '1' {
			count += 1
		} else if c == '0' {
			countZero += 1
		}
	}
	if count+countZero != 512 { // TODO: Hack
		tmplog.Println("bit count", count+countZero)
		panic("bit field is not 512")
	}

	return count
}
