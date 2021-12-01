package rpc

import (
	"context"
	"fmt"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/cache/depositcache"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	opfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/operation"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/db"
	lightservice "github.com/prysmaticlabs/prysm/beacon-chain/light"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/attestations"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/slashings"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/synccommittee"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/voluntaryexits"
	"github.com/prysmaticlabs/prysm/beacon-chain/p2p"
	"github.com/prysmaticlabs/prysm/beacon-chain/powchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/eth/debug"
	debugv1alpha1 "github.com/prysmaticlabs/prysm/beacon-chain/rpc/prysm/v1alpha1/debug"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/prysm/v1alpha1/light"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/statefetcher"
	slasherservice "github.com/prysmaticlabs/prysm/beacon-chain/slasher"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	chainSync "github.com/prysmaticlabs/prysm/beacon-chain/sync"
	"github.com/prysmaticlabs/prysm/config/features"
	"github.com/prysmaticlabs/prysm/monitoring/tracing"
	ethpbservice "github.com/prysmaticlabs/prysm/proto/eth/service"
	ethpbv1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	tmplog "log"
	"net"
	"sync"
	"time"
)

// headSyncMinEpochsAfterCheckpoint defines how many epochs should elapse after known finalization
// checkpoint for head sync to be triggered.
const headSyncMinEpochsAfterCheckpoint = 128

type Service struct {
	cfg *Config
	ctx context.Context

	cancel               context.CancelFunc
	listener             net.Listener
	grpcServer           *grpc.Server
	canonicalStateChan   chan *ethpbv1alpha1.BeaconState
	incomingAttestation  chan *ethpbv1alpha1.Attestation
	credentialError      error
	connectedRPCClients  map[net.Addr]bool
	clientConnectionLock sync.Mutex
}

type Config struct {
	Host                    string
	Port                    string
	CertFlag                string
	KeyFlag                 string
	BeaconMonitoringHost    string
	BeaconMonitoringPort    int
	BeaconDB                db.HeadAccessDatabase
	ChainInfoFetcher        blockchain.ChainInfoFetcher
	HeadFetcher             blockchain.HeadFetcher
	CanonicalFetcher        blockchain.CanonicalFetcher
	ForkFetcher             blockchain.ForkFetcher
	FinalizationFetcher     blockchain.FinalizationFetcher
	AttestationReceiver     blockchain.AttestationReceiver
	BlockReceiver           blockchain.BlockReceiver
	POWChainService         powchain.Chain
	ChainStartFetcher       powchain.ChainStartFetcher
	GenesisTimeFetcher      blockchain.TimeFetcher
	GenesisFetcher          blockchain.GenesisFetcher
	EnableDebugRPCEndpoints bool
	MockEth1Votes           bool
	AttestationsPool        attestations.Pool
	ExitPool                voluntaryexits.PoolManager
	SlashingsPool           slashings.PoolManager
	SlashingChecker         slasherservice.SlashingChecker
	SyncCommitteeObjectPool synccommittee.Pool
	SyncService             chainSync.Checker
	Broadcaster             p2p.Broadcaster
	PeersFetcher            p2p.PeersProvider
	PeerManager             p2p.PeerManager
	MetadataProvider        p2p.MetadataProvider
	DepositFetcher          depositcache.DepositFetcher
	PendingDepositFetcher   depositcache.PendingDepositsFetcher
	StateNotifier           statefeed.Notifier
	BlockNotifier           blockfeed.Notifier
	OperationNotifier       opfeed.Notifier
	StateGen                *stategen.State
	MaxMsgSize              int
	LightClientService      *lightservice.Service
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	svr := &Service{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
	return svr, nil
}

//// Start a blockchain service's main event loop.
//func (s *Service) Start() {
//	go s.process(s.ctx)
//}

func (s *Service) process(ctx context.Context) {
	count := 0
	for {
		tmplog.Println("rpc processing", count)
		time.Sleep(time.Second * 10)
		count += 1
	}
}

func (s *Service) Stop() error {
	defer s.cancel()
	return nil
}

func (s *Service) Status() error {
	return nil
}

// Start the gRPC server.
func (s *Service) Start() {
	address := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("Could not listen to port in Start() %s: %v", address, err)
	}
	s.listener = lis
	log.WithField("address", address).Info("gRPC server listening on port")
	tmplog.Println("address", address)

	opts := []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.StreamInterceptor(middleware.ChainStreamServer(
			recovery.StreamServerInterceptor(
				recovery.WithRecoveryHandlerContext(tracing.RecoveryHandlerFunc),
			),
			grpc_prometheus.StreamServerInterceptor,
			grpc_opentracing.StreamServerInterceptor(),
			s.validatorStreamConnectionInterceptor,
		)),
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandlerContext(tracing.RecoveryHandlerFunc),
			),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_opentracing.UnaryServerInterceptor(),
			s.validatorUnaryConnectionInterceptor,
		)),
		grpc.MaxRecvMsgSize(s.cfg.MaxMsgSize),
	}
	grpc_prometheus.EnableHandlingTimeHistogram()
	if s.cfg.CertFlag != "" && s.cfg.KeyFlag != "" {
		creds, err := credentials.NewServerTLSFromFile(s.cfg.CertFlag, s.cfg.KeyFlag)
		if err != nil {
			log.WithError(err).Fatal("Could not load TLS keys")
		}
		opts = append(opts, grpc.Creds(creds))
	} else {
		log.Warn("You are using an insecure gRPC server. If you are running your beacon node and " +
			"validator on the same machines, you can ignore this message. If you want to know " +
			"how to enable secure connections, see: https://docs.prylabs.network/docs/prysm-usage/secure-grpc")
	}
	s.grpcServer = grpc.NewServer(opts...)

	lightClientServer := &light.Server{
		LightClientService: s.cfg.LightClientService,
	}

	ethpbv1alpha1.RegisterLightClientServer(s.grpcServer, lightClientServer)
	if s.cfg.EnableDebugRPCEndpoints {
		log.Info("Enabled debug gRPC endpoints")
		debugServer := &debugv1alpha1.Server{
			GenesisTimeFetcher: s.cfg.GenesisTimeFetcher,
			BeaconDB:           s.cfg.BeaconDB,
			StateGen:           s.cfg.StateGen,
			HeadFetcher:        s.cfg.HeadFetcher,
			PeerManager:        s.cfg.PeerManager,
			PeersFetcher:       s.cfg.PeersFetcher,
		}
		debugServerV1 := &debug.Server{
			BeaconDB:    s.cfg.BeaconDB,
			HeadFetcher: s.cfg.HeadFetcher,
			StateFetcher: &statefetcher.StateProvider{
				BeaconDB:           s.cfg.BeaconDB,
				ChainInfoFetcher:   s.cfg.ChainInfoFetcher,
				GenesisTimeFetcher: s.cfg.GenesisTimeFetcher,
				StateGenService:    s.cfg.StateGen,
			},
		}
		ethpbv1alpha1.RegisterDebugServer(s.grpcServer, debugServer)
		ethpbservice.RegisterBeaconDebugServer(s.grpcServer, debugServerV1)
	}
	reflection.Register(s.grpcServer)

	go func() {
		if s.listener != nil {
			if err := s.grpcServer.Serve(s.listener); err != nil {
				log.Errorf("Could not serve gRPC: %v", err)
			}
		}
	}()
}

func (s *Service) validatorStreamConnectionInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	s.logNewClientConnection(ss.Context())
	return handler(srv, ss)
}

// Unary interceptor for new validator client connections to the beacon node.
func (s *Service) validatorUnaryConnectionInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.logNewClientConnection(ctx)
	return handler(ctx, req)
}

func (s *Service) logNewClientConnection(ctx context.Context) {
	if features.Get().DisableGRPCConnectionLogs {
		return
	}
	if clientInfo, ok := peer.FromContext(ctx); ok {
		// Check if we have not yet observed this grpc client connection
		// in the running beacon node.
		s.clientConnectionLock.Lock()
		defer s.clientConnectionLock.Unlock()
		if !s.connectedRPCClients[clientInfo.Addr] {
			log.WithFields(log.Fields{
				"addr": clientInfo.Addr.String(),
			}).Infof("New gRPC client connected to beacon node")
			s.connectedRPCClients[clientInfo.Addr] = true
		}
	}
}
