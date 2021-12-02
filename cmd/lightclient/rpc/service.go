package rpc

import (
	"context"
	"fmt"
	lightsync "github.com/prysmaticlabs/prysm/cmd/lightclient/sync"
	ethpbv1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
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
	Host        string
	Port        string
	CertFlag    string
	KeyFlag     string
	MaxMsgSize  int
	SyncService *lightsync.Service
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

	// GRPC Server
	opts := []grpc.ServerOption{}
	s.grpcServer = grpc.NewServer(opts...)

	// Register endpoints
	ethpbv1alpha1.RegisterLightNodeServer(s.grpcServer, &Server{
		syncService: s.cfg.SyncService,
	})
	reflection.Register(s.grpcServer)

	go func() {
		if s.listener != nil {
			if err := s.grpcServer.Serve(s.listener); err != nil {
				log.Errorf("Could not serve gRPC: %v", err)
			}
		}
	}()
}
