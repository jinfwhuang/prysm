package light

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prysmaticlabs/prysm/beacon-chain/light"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
)

type Server struct {
	LightClientService *light.Service
}

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

const (
	updatesResponseSize = 1
)

Updates(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*UpdateResponse, error)
SkipSyncUpdate(ctx context.Context, in *SkipSyncRequest, opts ...grpc.CallOption) (*SkipSyncRequest, error)


func (s *Server) Updates(ctx context.Context, _ *empty.Empty) (*ethpb.UpdatesResponse, error) {
	q := s.LightClientService.Queue
	size := updatesResponseSize
	if size > q.Len() {
		size = q.Len()
	}
	updates := make([]*ethpb.LightClientUpdate, size)
	_updates := q.Peek(size)
	for i := 0; i < size; i++ {
		updates[i] = _updates[i].(*ethpb.LightClientUpdate)
		//root, _ := updates[i].HashTreeRoot()
		//tmplog.Println(root)
	}
	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) SkipSyncUpdate(ctx context.Context, req *ethpb.SkipSyncRequest) (*ethpb.LightClientUpdate, error) {
	tmplog.Println("requesting for skip-sync-udpate", req)

	return &ethpb.LightClientUpdate{}, nil
}
