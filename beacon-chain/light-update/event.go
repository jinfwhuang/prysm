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
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	log "github.com/sirupsen/logrus"
	tmplog "log"
)

func (s *Service) GetChainHeadAndState(ctx context.Context) (block.SignedBeaconBlock, state.BeaconState, error) {
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

	st, err := s.cfg.Database.State(ctx, bytesutil.ToBytes32(root))
	if err != nil {
		return nil, nil, err
	}
	if st == nil || st.IsNil() {
		tmplog.Println(s.cfg.Database.HasState(ctx, bytesutil.ToBytes32(root)))
		return nil, nil, fmt.Errorf("head state is nil: blockRoot=%s", base64.StdEncoding.EncodeToString(root))
	}

	return blk, st, nil
}

func (s *Service) subscribeEvents(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.NewHead {
				tmplog.Println("head data")
				headEvent, ok := ev.Data.(*ethpbv1.EventHead)
				if !ok {
					panic(ok)
				}
				root := headEvent.Block

				//ev.Data
				//&ethpbv1.EventHead{
				//	Slot:                      newHeadSlot,
				//	Block:                     newHeadRoot,
				//	State:                     newHeadStateRoot,
				//	EpochTransition:           slots.IsEpochStart(newHeadSlot),
				//	PreviousDutyDependentRoot: previousDutyDependentRoot,
				//	CurrentDutyDependentRoot:  currentDutyDependentRoot,
				//}

				//root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
				//if err != nil {
				//	log.Error(err)
				//	continue
				//}
				//head, err := s.cfg.HeadFetcher.HeadBlock(ctx)
				////_head, err := head.Block
				//headRoot, err := head.Block().HashTreeRoot()
				//tmplog.Println(root)
				//tmplog.Println(headRoot)
				//if err != nil {
				//	log.Error(err)
				//	continue
				//}
				s.learnState(ctx, root[:])

				//head, beaconState, err := s.GetChainHeadAndState(ctx)
				//if err != nil {
				//	log.Error(err)
				//	continue
				//}
				//if err := s.processHeadEvent(ctx, head, beaconState); err != nil {
				//	log.Error(err)
				//	continue
				//}

			} else if ev.Type == statefeed.FinalizedCheckpoint {
				tmplog.Println("=================")
				tmplog.Println("=================")
				tmplog.Println("FinalizedCheckpoint")
				tmplog.Println("=================")
				tmplog.Println("=================")
				root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
				if err != nil {
					log.Error(err)
					continue
				}
				s.learnState(ctx, root)

				////head, beaconState, err := s.GetChainHeadAndState(ctx)
				//block, state, err := s.parseFinalizedEvent(ctx, ev.Data)
				//if err != nil {
				//	log.Error(err)
				//	continue
				//}
				//
				//if err := s.processFinalizedEvent(ctx, block, state); err != nil {
				//	tmplog.Println(err)
				//	log.Error(err)
				//	continue
				//}
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
	tmplog.Println("processFinalizedEvent")
	skipSyncUpdate, err := s.getFinalizedSkipSyncUpdate(ctx, block, state)
	tmplog.Println("created new update, ", skipSyncUpdate)
	if err != nil {
		return err
	}
	if skipSyncUpdate == nil {
		panic("creating a nil update")
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
	if newUpdate == nil {
		tmplog.Println("new update current sync", newUpdate)
	}

	key, err := newUpdate.CurrentSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	update, err := s.GetSkipSyncUpdate(ctx, bytesutil.ToBytes32(key[:]))
	if err != nil || update == nil {
		return newUpdate
	}

	tmplog.Println("old", update.SyncCommitteeBits)
	tmplog.Println("new", newUpdate.SyncCommitteeBits)

	oldParticipation := numOfSetBits(update.SyncCommitteeBits)
	newParticipation := numOfSetBits(newUpdate.SyncCommitteeBits)

	tmplog.Println("oldParticipation", oldParticipation, "newParticipation", newParticipation)

	//// Criteria 1: Compare which update has the most participation
	//if newParticipation >= oldParticipation {
	//	return newUpdate
	//} else {
	//	return update
	//}
	return newUpdate
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
		FinalityHeader:             &ethpb.BeaconBlockHeader{},
		FinalityBranch:             [][]byte{},
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
	if finalityBlock == nil {
		return nil, fmt.Errorf("cannot find block with root: %s", base64.StdEncoding.EncodeToString(fCheckpoint.Root))
	}
	tmplog.Println("found finality block", finalityBlock)
	tmplog.Println("bytesutil.ToBytes32(fCheckpoint.Root)", base64.StdEncoding.EncodeToString(fCheckpoint.Root))
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
	if err != nil {
		return nil, err
	}
	if finalizedState == nil {
		return nil, fmt.Errorf("cannot find state with root: %s", base64.StdEncoding.EncodeToString(fCheckpoint.Root))
	}
	tmplog.Println("found finality state", finalityBlock)
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

func (s *Service) parseFinalizedEvent(ctx context.Context, eventData interface{}) (block.SignedBeaconBlock, state.BeaconState, error) {
	finalizedCheckpoint, ok := eventData.(*v1.EventFinalizedCheckpoint)
	if !ok {
		return nil, nil, errors.New("expected finalized checkpoint event")
	}
	blk, err := s.getBlock(ctx, finalizedCheckpoint.Block)
	if err != nil {
		return nil, nil, err
	}
	st, err := s.getState(ctx, finalizedCheckpoint.Block)
	return blk, st, nil
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
