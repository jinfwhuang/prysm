package light

import (
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prysmaticlabs/prysm/beacon-chain/light"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
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
	}

	updateNextSyncCommRoot, _ := updates[0].NextSyncCommittee.HashTreeRoot()
	tmplog.Println("update      Next:", base64.StdEncoding.EncodeToString(updateNextSyncCommRoot[:]))

	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) GetSkipSyncUpdate(ctx context.Context, req *ethpb.SkipSyncRequest) (*ethpb.SkipSyncUpdate, error) {
	update, err := s.LightClientService.GetSkipSyncUpdate(ctx, bytesutil.ToBytes32(req.Key))
	if err != nil {
		return nil, err
	}

	currentSynRoot, _ := update.CurrentSyncCommittee.HashTreeRoot()
	nextSyncRoot, _ := update.NextSyncCommittee.HashTreeRoot()
	tmplog.Println("skip      Current:", base64.StdEncoding.EncodeToString(currentSynRoot[:]))
	tmplog.Println("skip         Next:", base64.StdEncoding.EncodeToString(nextSyncRoot[:]))

	return update, nil
}
