package light

import (
	"context"
	tmplog "log"

	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

type Server struct {
	//Database iface.LightClientDatabase

}

//type LightClientServer interface {
//	Updates(context.Context, *empty.Empty) (*UpdatesResponse, error)
//	SkipSyncUpdate(context.Context, *SkipsyncRequest) (*LightClientUpdate, error)
//}
func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func (s *Server) Updates(ctx context.Context, _ *empty.Empty) (*ethpb.UpdatesResponse, error) {
	tmplog.Println("requesting for updates")
	updates := make([]*ethpb.LightClientUpdate, 0)
	//for _, period := range req.SyncCommitteePeriods {
	//	update, err := s.Database.LightClientBestUpdateForPeriod(ctx, period)
	//	if errors.Is(err, kv.ErrNotFound) {
	//		continue
	//	} else if err != nil {
	//		return nil, status.Errorf(codes.Internal, "Could not retrieve best update for %d: %v", period, err)
	//	}
	//	updates = append(updates, update)
	//}
	return &ethpb.UpdatesResponse{Updates: updates}, nil
}

func (s *Server) SkipSyncUpdate(ctx context.Context, req *ethpb.SkipsyncRequest) (*ethpb.LightClientUpdate, error) {
	tmplog.Println("requesting for skip-sync-udpate", req)

	return &ethpb.LightClientUpdate{}, nil
}
