package light

import (
	"context"
	"encoding/base64"
	"encoding/hex"
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

func (s *Server) GetUpdates(ctx context.Context, _ *empty.Empty) (*ethpb.UpdatesResponse, error) {
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

func (s *Server) GetSkipSyncUpdate(ctx context.Context, req *ethpb.SkipSyncRequest) (*ethpb.SkipSyncUpdate, error) {
	//tmplog.Println("requesting for skip-sync-udpate", req)
	tmplog.Println("-----------")
	tmplog.Println("hex    ", hex.EncodeToString(req.Key))
	tmplog.Println("base64 ", base64.StdEncoding.EncodeToString(req.Key))

	return &ethpb.SkipSyncUpdate{}, nil
}
