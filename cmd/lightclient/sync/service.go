package sync

import (
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/ptypes/empty"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcutil "github.com/prysmaticlabs/prysm/api/grpc"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	tmplog "log"
	"sync"
	"time"
)

//type Store struct {
//	Snapshot     *ethpb.ClientSnapshot
//	ValidUpdates []*ethpb.LightClientUpdate
//}

type Service struct {
	cfg    *Config
	ctx    context.Context
	cancel context.CancelFunc
	lock   sync.RWMutex

	dataDir string
	store   *ethpb.LightClientStore

	// grpc
	conn *grpc.ClientConn
	//grpcRetryDelay        time.Duration
	//grpcRetries           uint
	//maxCallRecvMsgSize    int
	//withCert              string
	//endpoint              string

	lightClientServer ethpb.LightClientClient
}

type Config struct {
	// The root is used if there is no local DB of LightClientStore
	TrustedCurrentCommitteeRoot string // Merkle root in base64 string
	GrpcRetryDelay              time.Duration
	GrpcRetries                 int
	MaxCallRecvMsgSize          int
	GrpcEndpoint                string
}

/**
Design:
1. The service keep track of "store"
2. The service save the latest store to disk

3. The service recove from "store"
*/

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)

	svr := &Service{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
	return svr, nil
}

// Start a blockchain service's main event loop.
func (s *Service) Start() {
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(s.cfg.MaxCallRecvMsgSize),
			grpc_retry.WithMax(uint(s.cfg.GrpcRetries)),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(s.cfg.GrpcRetryDelay)),
		),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
			grpc_opentracing.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
			grpc_retry.UnaryClientInterceptor(),
			grpcutil.LogRequests,
		)),
		grpc.WithChainStreamInterceptor(
			grpcutil.LogStream,
			grpc_opentracing.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
			grpc_retry.StreamClientInterceptor(),
		),
		grpc.WithInsecure(),
		//grpc.WithResolvers(&multipleEndpointsGrpcResolverBuilder{}),
	}
	//v.ctx = grpcutil.AppendHeaders(s.ctx, s.grpcHeaders)

	conn, err := grpc.DialContext(s.ctx, s.cfg.GrpcEndpoint, dialOpts...)
	if err != nil {
		panic(err)
	}
	s.conn = conn
	s.lightClientServer = ethpb.NewLightClientClient(s.conn)

	s.init(s.ctx)
	tmplog.Println(s.store)
	go s.process(s.ctx)
}

func (s *Service) init(ctx context.Context) {
	// 1. Recover from a file location if possible

	// 2. Use hard coded trustedCurrentCommitteeRoot
	s.skipSync(ctx)
}

// Skip sync to the point where this client could use the current LightClientUpdates
func (s *Service) skipSync(ctx context.Context) {
	// Find our what is the expected  next-sync-committee
	resp, err := s.lightClientServer.GetUpdates(ctx, &empty.Empty{})
	if err != nil || len(resp.GetUpdates()) < 1 {
		panic(err) // TODO: fix
	}
	update := resp.Updates[0]
	nextSyncComm := update.NextSyncCommittee
	expectedNextSyncCommRoot, err := nextSyncComm.HashTreeRoot()
	if err != nil {
		panic(err)
	}

	// Skip sync until we could use the expected next-sync-committee
	for {
		skipsyncKey, err := base64.StdEncoding.DecodeString(s.cfg.TrustedCurrentCommitteeRoot)
		if err != nil {
			panic(err)
		}
		skipsyncUpdate, err := s.lightClientServer.GetSkipSyncUpdate(ctx, &ethpb.SkipSyncRequest{
			Key: skipsyncKey,
		})
		//validateSkipSyncUpdate(skipsyncResp)
		// TODO apply the update
		s.processSkipSyncUpdate(skipsyncUpdate)
		storedNextSyncCommRoot, err := s.store.Store.NextSyncCommittee.HashTreeRoot()

		tmplog.Println("expected Next: ", base64.StdEncoding.EncodeToString(expectedNextSyncCommRoot[:]))
		tmplog.Println("stored Next:   ", base64.StdEncoding.EncodeToString(storedNextSyncCommRoot[:]))

		if err != nil {
			panic(err)
		}
		if storedNextSyncCommRoot == expectedNextSyncCommRoot {
			break
		}
	}

}

func (s *Service) processSkipSyncUpdate(update *ethpb.SkipSyncUpdate) {
	lightClientUpdate := &ethpb.LightClientUpdate{
		Header:                  update.Header,
		NextSyncCommittee:       update.NextSyncCommittee,
		NextSyncCommitteeBranch: update.NextSyncCommitteeBranch,
		FinalityHeader:          update.FinalityHeader,
		FinalityBranch:          update.FinalityBranch,
		SyncCommitteeBits:       update.SyncCommitteeBits,
		SyncCommitteeSignature:  update.SyncCommitteeSignature,
		ForkVersion:             update.ForkVersion,
	}
	s.processLightClientUpdate(lightClientUpdate)
}

func (s *Service) processLightClientUpdate(update *ethpb.LightClientUpdate) {
	processLightClientUpdate(s.store, update, s.store.Store.Header.Slot, GenesisValidatorsRoot)
}

//func (s *Service) process(ctx context.Context) {
//
//	count := 0
//	for {
//
//		tmplog.Println("processing", count)
//		time.Sleep(time.Second * 10)
//		count += 1
//	}
//	tmplog.Println("xxx light client sync done processing xxx")
//}

// Stop the blockchain service's main event loop and associated goroutines.
func (s *Service) Stop() error {
	defer s.cancel()
	return nil
}

// Status always returns nil unless there is an error condition that causes
// this service to be unhealthy.
func (s *Service) Status() error {
	return nil
}
