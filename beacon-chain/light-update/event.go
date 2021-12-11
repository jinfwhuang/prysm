package light

import "C"
import (
	"context"
	"encoding/base64"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
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
				s.processFinalizedEvent(ctx, root)
			}
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) processHeadEvent(ctx context.Context, root []byte) error {
	skipSyncUpdate, err := s.getNonFinalizedSkipSyncUpdate(ctx, root)
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
func (s *Service) processFinalizedEvent(ctx context.Context, root []byte) error {
	skipSyncUpdate, err := s.getFinalizedSkipSyncUpdate(ctx, root)
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
