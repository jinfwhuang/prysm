package main

import (
	"context"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcutil "github.com/prysmaticlabs/prysm/api/grpc"
	lightnode "github.com/prysmaticlabs/prysm/cmd/lightclient/node"
	"github.com/prysmaticlabs/prysm/proto/eth/service"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	tmplog "log"
	"os"

	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	v2 "github.com/prysmaticlabs/prysm/proto/eth/v2"
)

func main() {
	ctx := context.Background()
	dialOpts := []grpc.DialOption{
		//grpc.WithDefaultCallOptions(
		//	grpc.MaxCallRecvMsgSize(s.cfg.GrpcMaxCallRecvMsgSize),
		//	grpc_retry.WithMax(uint(s.cfg.GrpcRetries)),
		//	grpc_retry.WithBackoff(grpc_retry.BackoffLinear(s.cfg.GrpcRetryDelay)),
		//),
		//grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		//grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
		//	grpc_opentracing.UnaryClientInterceptor(),
		//	grpc_prometheus.UnaryClientInterceptor,
		//	grpc_retry.UnaryClientInterceptor(),
		//	grpcutil.LogRequests,
		//)),
		//grpc.WithChainStreamInterceptor(
		//	grpcutil.LogStream,
		//	grpc_opentracing.StreamClientInterceptor(),
		//	grpc_prometheus.StreamClientInterceptor,
		//	grpc_retry.StreamClientInterceptor(),
		//),
		grpc.WithInsecure(),
		//grpc.WithResolvers(&multipleEndpointsGrpcResolverBuilder{}),
	}
	//v.ctx = grpcutil.AppendHeaders(s.ctx, s.grpcHeaders)

	tmplog.Printf("connecting to grpc server: %s", "localhost:4000")
	conn, err := grpc.DialContext(ctx, "localhost:4000", dialOpts...)
	if err != nil {
		panic(err)
	}
	server1 := service.NewBeaconChainClient(conn)

	headers, err := server1.ListBlockHeaders(ctx, &v1.BlockHeadersRequest{})
	if err != nil {
		panic(err)
	}
	data := headers.Data[0]
	stateRoot := data.Header.Message.StateRoot

	server2 := ethpb.NewBeaconChainClient(conn)
	server2.GetChainHead()

	resp, err := server1.ListSyncCommittees(ctx, &v2.StateSyncCommitteesRequest{
		StateId: stateRoot,
	})

	ethpb.SyncCommittee{}

	if err != nil {
		panic(err)
	}
	resp.Data

}
