package light

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"

	//vv1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	statev2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	//"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	//"github.com/prysmaticlabs/prysm/encoding/ssz"
	//"github.com/prysmaticlabs/prysm/network/forks"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
)

const (
	finalizedCheckpointStateIndex = 20
	nextSyncCommitteeStateIndex   = 23
)

func (s *Service) getChainHeadAndState(ctx context.Context) (block.SignedBeaconBlock, state.BeaconState, error) {
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

func (s *Service) subscribeHeadEvent(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.NewHead {
				head, beaconState, err := s.getChainHeadAndState(ctx)
				if err != nil {
					log.Error(err)
					continue
				}
				if err := s.maintainQueueLightClientUpdates(ctx, head, beaconState); err != nil {
					log.Error(err)
					continue
				}
			}
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

// Use the blocks to build light-client-updates
func (s *Service) maintainQueueLightClientUpdates(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) error {
	_state := state.(*statev2.BeaconState)
	ztypeState := ztype.FromBeaconState(_state)

	// Header
	header, err := block.Header()
	if err != nil {
		return err
	}

	//currentCom := state.CurrentSyncCommittee()
	nextCom, err := state.NextSyncCommittee()
	if err != nil {
		return err
	}
	nextSyncCommitteeBranch := ztypeState.GetBranch(ztypeState.GetGIndex(23))

	// Finality information
	fCheckpoint := state.FinalizedCheckpoint()
	finalityBlock, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return err
	}
	signedFinalityHeader, err := finalityBlock.Header()
	finalityHeader := signedFinalityHeader.Header
	if err != nil {
		return err
	}
	finalityCheckpointRootBranch := ztypeState.GetBranch(ztypeState.GetGIndex(20, 1))

	syncAgg, err := block.Block().Body().SyncAggregate()
	if err != nil {
		return err
	}

	update := &ethpb.LightClientUpdate{
		Header:                  header.Header,
		NextSyncCommittee:       nextCom,
		NextSyncCommitteeBranch: nextSyncCommitteeBranch,
		FinalityHeader:          finalityHeader,
		FinalityBranch:          finalityCheckpointRootBranch,
		SyncCommitteeBits:       syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature:  syncAgg.SyncCommitteeSignature,
		ForkVersion:             state.Fork().CurrentVersion,
	}

	s.Queue.Enqueue(update)

	tmplog.Println("light-client-update queue", s.Queue.Len(), s.Queue.Cap())
	//saveSsz(block.Block(), ztypeState, finalityBlock.Block())

	return nil
}

// Use the blocks to build light-client-updates
func (s *Service) maintainSkipSyncUpdates(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconState) error {
	_state := state.(*statev2.BeaconState)
	ztypeState := ztype.FromBeaconState(_state)

	// Header
	header, err := block.Header()
	if err != nil {
		return err
	}

	//currentCom := state.CurrentSyncCommittee()
	nextCom, err := state.NextSyncCommittee()
	if err != nil {
		return err
	}
	nextSyncCommitteeBranch := ztypeState.GetBranch(ztypeState.GetGIndex(23))

	// Finality information
	fCheckpoint := state.FinalizedCheckpoint()
	finalityBlock, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return err
	}
	signedFinalityHeader, err := finalityBlock.Header()
	finalityHeader := signedFinalityHeader.Header
	if err != nil {
		return err
	}
	finalityCheckpointRootBranch := ztypeState.GetBranch(ztypeState.GetGIndex(20, 1))

	syncAgg, err := block.Block().Body().SyncAggregate()
	if err != nil {
		return err
	}

	update := &ethpb.LightClientUpdate{
		Header:                  header.Header,
		NextSyncCommittee:       nextCom,
		NextSyncCommitteeBranch: nextSyncCommitteeBranch,
		FinalityHeader:          finalityHeader,
		FinalityBranch:          finalityCheckpointRootBranch,
		SyncCommitteeBits:       syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature:  syncAgg.SyncCommitteeSignature,
		ForkVersion:             state.Fork().CurrentVersion,
	}

	s.Queue.Enqueue(update)

	tmplog.Println("light-client-update queue", s.Queue.Len(), s.Queue.Cap())
	//saveSsz(block.Block(), ztypeState, finalityBlock.Block())

	return nil
}
