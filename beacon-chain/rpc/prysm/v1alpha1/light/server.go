package light

import (
	"context"
	"github.com/prysmaticlabs/prysm/beacon-chain/light"
	tmplog "log"
	"math/rand"

	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

type Server struct {
	LightClientService *light.Service
}

//type LightClientServer interface {
//	Updates(context.Context, *empty.Empty) (*UpdatesResponse, error)
//	SkipSyncUpdate(context.Context, *SkipsyncRequest) (*LightClientUpdate, error)
//}
func init() {
	tmplog.SetFlags(tmplog.Llongfile)

}

func (s *Server) Updates(ctx context.Context, _ *empty.Empty) (*ethpb.UpdatesResponse, error) {
	q := s.LightClientService.Queue
	size := rand.Intn(q.Len())
	if size == 0 && q.Len() > 0 {
		size = 1
	}
	tmplog.Println("requesting size", size, "q size", q.Len(), "q cap", q.Cap())

	updates := make([]*ethpb.LightClientUpdate, size)
	_updates := q.Peek(size)
	for i := 0; i < size; i++ {
		updates[i] = _updates[i].(*ethpb.LightClientUpdate)
	}
	tmplog.Println("found updats", len(updates))
	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) SkipSyncUpdate(ctx context.Context, req *ethpb.SkipsyncRequest) (*ethpb.LightClientUpdate, error) {
	tmplog.Println("requesting for skip-sync-udpate", req)

	return &ethpb.LightClientUpdate{}, nil
}
