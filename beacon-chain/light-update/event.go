package light

import "C"
import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	statev1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	statev2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	log "github.com/sirupsen/logrus"
	tmplog "log"
)

func (s *Service) GetChainHeadAndState(ctx context.Context) (block.SignedBeaconBlock, state.BeaconState, error) {
	head, err := s.cfg.HeadFetcher.HeadBlock(ctx)
	if err != nil {
		return nil, nil, err
	}
	if head == nil || head.IsNil() {
		return nil, nil, errors.New("head block is nil")
	}
	st, err := s.cfg.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, nil, errors.New("head state is nil")
	}
	if st == nil || st.IsNil() {
		return nil, nil, err
	}
	return head, st, nil
}

func (s *Service) subscribeEvents(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.NewHead || ev.Type == statefeed.FinalizedCheckpoint {
				head, beaconState, err := s.GetChainHeadAndState(ctx)
				if err != nil {
					log.Error(err)
					continue
				}
				if ev.Type == statefeed.NewHead {
					if err := s.processHeadEvent(ctx, head, beaconState); err != nil {
						log.Error(err)
						continue
					}
				} else if ev.Type == statefeed.FinalizedCheckpoint {
					if err := s.processFinalizedEvent(ctx, head, beaconState); err != nil {
						log.Error(err)
						continue
					}
				}
			}
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) processHeadEvent(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) error {
	skipSyncUpdate, err := s.getNonFinalizedSkipSyncUpdate(ctx, block, state)
	if err != nil {
		return err
	}
	if skipSyncUpdate.CurrentSyncCommittee == nil {
		panic("dfdfdfdf")
	}

	skipSyncUpdate = s.bestSkipSyncUpdate(ctx, skipSyncUpdate)
	update := ToLightClientUpdate(skipSyncUpdate)

	// Save data
	s.Queue.Enqueue(update)
	s.cfg.Database.SaveSkipSyncUpdate(ctx, skipSyncUpdate)

	// DEBUG
	currentSyncCommRoot, err := skipSyncUpdate.CurrentSyncCommittee.HashTreeRoot()
	nextSyncCommRoot, err := skipSyncUpdate.NextSyncCommittee.HashTreeRoot()
	tmplog.Println("Current :", base64.StdEncoding.EncodeToString(currentSyncCommRoot[:]))
	tmplog.Println("Next    :", base64.StdEncoding.EncodeToString(nextSyncCommRoot[:]))

	return nil
}

func (s *Service) processFinalizedEvent(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) error {
	skipSyncUpdate, err := s.getFinalizedSkipSyncUpdate(ctx, block, state)
	if err != nil {
		return err
	}
	skipSyncUpdate = s.bestSkipSyncUpdate(ctx, skipSyncUpdate)
	update := ToLightClientUpdate(skipSyncUpdate)

	// Save data
	s.Queue.Enqueue(update)
	s.cfg.Database.SaveSkipSyncUpdate(ctx, skipSyncUpdate)

	// DEBUG
	currentSyncCommRoot, err := skipSyncUpdate.CurrentSyncCommittee.HashTreeRoot()
	nextSyncCommRoot, err := skipSyncUpdate.NextSyncCommittee.HashTreeRoot()
	tmplog.Println("Current :", base64.StdEncoding.EncodeToString(currentSyncCommRoot[:]))
	tmplog.Println("Next    :", base64.StdEncoding.EncodeToString(nextSyncCommRoot[:]))

	return nil
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

	tmplog.Println(update.SyncCommitteeBits)
	tmplog.Println(newUpdate.SyncCommitteeBits)

	oldParticipation := numOfSetBits(update.SyncCommitteeBits)
	newParticipation := numOfSetBits(newUpdate.SyncCommitteeBits)

	tmplog.Println("oldParticipation", oldParticipation, "newParticipation", newParticipation)

	// Criteria 1: Compare which update has the most participation
	if newParticipation >= oldParticipation {
		return newUpdate
	} else {
		return update
	}
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
	//if count+countZero != 512 { // TODO: Hack
	//	tmplog.Println("bit count", count+countZero)
	//	panic("bit field is not 512")
	//}

	return count
}

func (s *Service) getNonFinalizedSkipSyncUpdate(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) (*ethpb.SkipSyncUpdate, error) {
	attestedHeader, err := block.Header()
	if err != nil {
		return nil, err
	}

	switch v := state.(type) {
	case *statev1.BeaconState:
		return nil, fmt.Errorf("wrong type: interface conversion: state.BeaconState is *v1.BeaconState, not *v2.BeaconState")
	case *statev2.BeaconState:
		// nothing
	default:
		tmplog.Println("unknown type")
		tmplog.Println(v)
		return nil, fmt.Errorf("unknown type %s", v)
	}
	zState := ztype.FromBeaconState(state.(*statev2.BeaconState))

	syncAgg, err := block.Block().Body().SyncAggregate()
	if err != nil {
		return nil, err
	}

	currentComm, err := state.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	currentCommBranch := zState.GetBranch(zState.GetGIndex(22))

	nextComm, err := state.NextSyncCommittee()
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
		FinalityHeader:             nil,
		FinalityBranch:             nil,
		SyncCommitteeBits:          syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature:     syncAgg.SyncCommitteeSignature,
		ForkVersion:                state.Fork().CurrentVersion,
	}

	return update, nil
}

func (s *Service) getFinalizedSkipSyncUpdate(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) (*ethpb.SkipSyncUpdate, error) {
	// Based on state
	attestedHeader, err := block.Header()
	if err != nil {
		return nil, err
	}

	fCheckpoint := state.FinalizedCheckpoint()
	finalityBlock, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return nil, err
	}
	signedFinalityHeader, err := finalityBlock.Header()
	finalityHeader := signedFinalityHeader.Header
	if err != nil {
		return nil, err
	}
	zState := ztype.FromBeaconState(state.(*statev2.BeaconState))
	finalityCheckpointRootBranch := zState.GetBranch(zState.GetGIndex(20, 1))

	syncAgg, err := block.Block().Body().SyncAggregate()
	if err != nil {
		return nil, err
	}

	// Committee information refers to finalize_state
	finalizedState, err := s.cfg.Database.State(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil || finalizedState == nil {
		return nil, err
	}
	zFinalizedState := ztype.FromBeaconState(finalizedState.(*statev2.BeaconState))
	currentCom, err := finalizedState.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	currentSyncCommitteeBranch := zFinalizedState.GetBranch(zFinalizedState.GetGIndex(22))

	nextCom, err := finalizedState.NextSyncCommittee()
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
		ForkVersion:                state.Fork().CurrentVersion,
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
