package rpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	lightsync "github.com/prysmaticlabs/prysm/cmd/lightclient/sync"

	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
)

type Server struct {
	//LightClientService *light.Service
	syncService *lightsync.Service
}

func (s *Server) Head(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	tmplog.Println("asking for head")
	tmplog.Println(s.syncService)
	return &ethpb.BeaconBlockHeader{}, nil
}

func (s *Server) FinalizedHead(ctx context.Context, _ *empty.Empty) (*ethpb.BeaconBlockHeader, error) {
	return &ethpb.BeaconBlockHeader{}, nil
}
