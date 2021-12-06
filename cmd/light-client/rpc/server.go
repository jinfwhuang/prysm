package rpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	lightsync "github.com/prysmaticlabs/prysm/cmd/light-client/sync"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
)

type Server struct {
	syncService *lightsync.Service
}

func (s *Server) debug() {
	snapshot := s.syncService.Store.Snapshot
	currentSysCommRoot, _ := snapshot.CurrentSyncCommittee.HashTreeRoot()
	nextSyncCommRoot, _ := snapshot.NextSyncCommittee.HashTreeRoot()

	tmplog.Println("--------------------------")
	tmplog.Println(snapshot.Header)
	tmplog.Println("period        :", int(snapshot.Header.Slot)/int(lightsync.EpochsPerSyncCommitteePeriod))
	tmplog.Println("Store  Current:", base64.StdEncoding.EncodeToString(currentSysCommRoot[:]))
	tmplog.Println("Store  Next   :", base64.StdEncoding.EncodeToString(nextSyncCommRoot[:]))
	tmplog.Println("--------------------------")
}

func (s *Server) Head(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	//s.debug()
	return s.syncService.Store.Snapshot.Header, nil
}

func (s *Server) FinalizedHead(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	return nil, fmt.Errorf("not implemented")
}
