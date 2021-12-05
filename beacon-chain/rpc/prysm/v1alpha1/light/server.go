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
	tmplog.Printf("GetUpdate, Next:%s", base64.StdEncoding.EncodeToString(updateNextSyncCommRoot[:]))
	tmplog.Printf("GetUpdate, Slot:%d", updates[0].Header.Slot)

	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) GetSkipSyncUpdate(ctx context.Context, req *ethpb.SkipSyncRequest) (*ethpb.SkipSyncUpdate, error) {
	update, err := s.LightClientService.GetSkipSyncUpdate(ctx, bytesutil.ToBytes32(req.Key))
	if err != nil {
		return nil, err
	}

	currentSynRoot, _ := update.CurrentSyncCommittee.HashTreeRoot()
	nextSyncRoot, _ := update.NextSyncCommittee.HashTreeRoot()
	tmplog.Printf("GetSkip, Current:%s", base64.StdEncoding.EncodeToString(currentSynRoot[:]))
	tmplog.Printf("GetSkip,    Next:%s", base64.StdEncoding.EncodeToString(nextSyncRoot[:]))
	tmplog.Printf("GetUpdate,  Slot:%d", update.Header.Slot)
	return update, nil
}

func (s *Server) DebugGetTrustedCurrentCommitteeRoot(ctx context.Context, empty *empty.Empty) (*ethpb.SkipSyncUpdate, error) {
	q := s.LightClientService.Queue
	updates := q.Peek(1)

	// Generate proto and put this
}

//rpc DebugGetTrustedCurrentCommitteeRoot(google.protobuf.Empty) returns (DebugGetTrustedCurrentCommitteeRootResp) {
//option (google.api.http) = {
//get: "/eth/v1alpha1/light-client/debug-get-trusted-sync-comm-root",
//};
//}
