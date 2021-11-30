package light

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz"
	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	"github.com/prysmaticlabs/prysm/time/slots"
)

const (
	finalizedCheckpointStateIndex = 20
	nextSyncCommitteeStateIndex   = 23
)

func (s *Service) subscribeFinalizedEvent(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.FinalizedCheckpoint {
				blk, beaconState, err := s.parseFinalizedEvent(ctx, ev.Data)
				if err != nil {
					log.Error(err)
					continue
				}
				if err := s.onFinalized(ctx, blk, beaconState); err != nil {
					log.Error(err)
					continue
				}
			}
		}
	}
}

func (s *Service) parseFinalizedEvent(
	ctx context.Context, eventData interface{},
) (block.SignedBeaconBlock, state.BeaconState, error) {
	finalizedCheckpoint, ok := eventData.(*v1.EventFinalizedCheckpoint)
	if !ok {
		return nil, nil, errors.New("expected finalized checkpoint event")
	}
	checkpointRoot := bytesutil.ToBytes32(finalizedCheckpoint.Block)
	blk, err := s.cfg.Database.Block(ctx, checkpointRoot)
	if err != nil {
		return nil, nil, err
	}
	if blk == nil || blk.IsNil() {
		return nil, nil, err
	}
	st, err := s.cfg.StateGen.StateByRoot(ctx, checkpointRoot)
	if err != nil {
		return nil, nil, err
	}
	if st == nil || st.IsNil() {
		return nil, nil, err
	}
	return blk, st, nil
}

func (s *Service) onFinalized(
	ctx context.Context, signedBlock block.SignedBeaconBlock, postState state.BeaconStateAltair,
) error {
	if _, ok := postState.InnerStateUnsafe().(*ethpb.BeaconStateAltair); !ok {
		return errors.New("expected an Altair beacon state")
	}
	blk := signedBlock.Block()
	header, err := block.BeaconBlockHeaderFromBlockInterface(blk)
	if err != nil {
		return err
	}
	tb, err := ssz.NewTreeBackedState(postState)
	if err != nil {
		return err
	}
	proof, gIndex, err := tb.Proof(nextSyncCommitteeStateIndex)
	if err != nil {
		return err
	}
	nextSyncCommittee, err := postState.NextSyncCommittee()
	if err != nil {
		return err
	}
	root, err := postState.HashTreeRoot(ctx)
	if err != nil {
		return err
	}
	nextSyncCommitteeRoot, err := nextSyncCommittee.HashTreeRoot()
	if err != nil {
		return err
	}
	log.Info("On finalized update")
	log.Infof("Header state root %#x, state hash tree root %#x", header.StateRoot, root)
	log.Infof("Generating proof against root %#x with gindex %d and leaf root %#x", root, gIndex, nextSyncCommitteeRoot)
	log.Info("-----")
	log.Infof("Proof with length %d", len(proof))
	for _, elem := range proof {
		log.Infof("%#x", bytesutil.Trunc(elem))
	}
	log.Info("-----")

	s.lock.Lock()
	defer s.lock.Unlock()
	currentEpoch := slots.ToEpoch(blk.Slot())
	s.finalizedByEpoch[currentEpoch] = &ethpb.LightClientFinalizedCheckpoint{
		Header:                  header,
		NextSyncCommittee:       nextSyncCommittee,
		NextSyncCommitteeBranch: proof,
	}
	return nil
}
