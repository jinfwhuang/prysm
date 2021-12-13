package light

import "C"
import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	tmplog "log"
)

func (s *Service) subscribeEvents(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.NewHead {
				headEvent, ok := ev.Data.(*ethpbv1.EventHead)
				if !ok {
					log.Error("cannot convert data to EventHead")
					log.Error(ok)
					continue
				}
				root := headEvent.Block
				s.processHeadEvent(ctx, root)
			} else if ev.Type == statefeed.FinalizedCheckpoint {
				tmplog.Println("working with a statefeed.FinalizedCheckpoint")
				root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
				if err != nil {
					log.Error(err)
					continue
				}
				s.processFinalizedEvent(ctx, root)
			}
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

// Save LightClientUpate
func (s *Service) processHeadEvent(ctx context.Context, root []byte) error {
	skipSyncUpdate, err := s.getNonFinalizedSkipSyncUpdate(ctx, root)
	if err != nil {
		return err
	}
	update := ToLightClientUpdate(skipSyncUpdate)

	// Save data
	s.Queue.Enqueue(update)

	// DEBUG
	currentSyncCommRoot, err := skipSyncUpdate.CurrentSyncCommittee.HashTreeRoot()
	nextSyncCommRoot, err := skipSyncUpdate.NextSyncCommittee.HashTreeRoot()
	tmplog.Println("Current :", base64.StdEncoding.EncodeToString(currentSyncCommRoot[:]))
	tmplog.Println("Next    :", base64.StdEncoding.EncodeToString(nextSyncCommRoot[:]))

	return nil
}

func IsEmptyHeader(header *ethpb.BeaconBlockHeader) bool {
	emptyHeader := &ethpb.BeaconBlockHeader{}
	return proto.Equal(header, emptyHeader)
}

// Save LightClientUpdate and SkipSyncUpdate
func (s *Service) processFinalizedEvent(ctx context.Context, root []byte) error {
	skipSyncUpdate, err := s.getFinalizedSkipSyncUpdate(ctx, root)
	if err != nil {
		return err
	}
	if skipSyncUpdate == nil {
		panic("creating a nil update")
	}
	if IsEmptyHeader(skipSyncUpdate.FinalityHeader) {
		return fmt.Errorf("header is empty") // Ensure that SkipSyncUpdate is only created from FinalizedEvent
	}
	_skipSyncUpdate := s.bestSkipSyncUpdate(ctx, skipSyncUpdate)
	if !IsEmptyHeader(_skipSyncUpdate.FinalityHeader) {
		skipSyncUpdate = _skipSyncUpdate
	}

	update := ToLightClientUpdate(skipSyncUpdate)

	if IsEmptyHeader(skipSyncUpdate.FinalityHeader) {
		return fmt.Errorf("header is empty")
	}

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
