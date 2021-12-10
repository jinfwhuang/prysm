package light

import (
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prysmaticlabs/prysm/beacon-chain/light-update"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
)

type Server struct {
	LightUpdateService *light.Service
}

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

const (
	updatesResponseSize = 1
)

func (s *Server) GetUpdates(ctx context.Context, _ *empty.Empty) (*ethpb.UpdatesResponse, error) {
	q := s.LightUpdateService.Queue
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
	tmplog.Printf("GetUpdate, Next:%s", base64.StdEncoding.EncodeToString(updateNextSyncCommRoot[:]))
	tmplog.Printf("GetUpdate, Slot:%d", updates[0].AttestedHeader.Slot)

	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) GetSkipSyncUpdate(ctx context.Context, req *ethpb.SkipSyncRequest) (*ethpb.SkipSyncUpdate, error) {
	update, err := s.LightUpdateService.GetSkipSyncUpdate(ctx, bytesutil.ToBytes32(req.Key))
	if err != nil {
		return nil, err
	}

	currentSynRoot, _ := update.CurrentSyncCommittee.HashTreeRoot()
	nextSyncRoot, _ := update.NextSyncCommittee.HashTreeRoot()
	tmplog.Printf("GetSkip, Current:%s", base64.StdEncoding.EncodeToString(currentSynRoot[:]))
	tmplog.Printf("GetSkip,    Next:%s", base64.StdEncoding.EncodeToString(nextSyncRoot[:]))
	tmplog.Printf("GetUpdate,  Slot:%d", update.AttestedHeader.Slot)
	return update, nil
}

// TODO: only use for debug only
func (s *Server) DebugGetTrustedCurrentCommitteeRoot(ctx context.Context, empty *empty.Empty) (*ethpb.DebugGetTrustedCurrentCommitteeRootResp, error) {
	syncComm, err := s.LightUpdateService.GetCurrentSyncComm(ctx)
	root, err := syncComm.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	return &ethpb.DebugGetTrustedCurrentCommitteeRootResp{
		Key: root[:],
	}, nil
}
