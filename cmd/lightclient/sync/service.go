package sync

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcutil "github.com/prysmaticlabs/prysm/api/grpc"
	"github.com/prysmaticlabs/prysm/config/params"
	"os"
	"path/filepath"
	"strings"

	//lightnode "github.com/prysmaticlabs/prysm/cmd/lightclient/node"
	lightclientdebug "github.com/prysmaticlabs/prysm/cmd/lightclient/debug"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	tmplog "log"
	"sync"
	"time"
)

//type Store struct {
//	Snapshot     *ethpb.ClientSnapshot
//	ValidUpdates []*ethpb.LightClientUpdate
//}

type SyncMode byte

const (
	ModeLatest SyncMode = iota
	ModeFinality
)

var (
	EpochsPerSyncCommitteePeriod = params.BeaconConfig().EpochsPerSyncCommitteePeriod
)

func (s SyncMode) String() string {
	return []string{
		"latest",
		"finality",
	}[s]
}

type Service struct {
	cfg    *Config
	ctx    context.Context
	cancel context.CancelFunc
	lock   sync.RWMutex

	dataDir string
	Store   *ethpb.LightClientStore

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
	TrustedCurrentCommitteeRoot string // Merkle root in base64 string
	SyncMode                    SyncMode
	DataDir                     string
	// grpc
	FullNodeServerEndpoint string
	GrpcMaxCallRecvMsgSize int
	GrpcRetryDelay         time.Duration
	GrpcRetries            int
}

/**
Design:
1. The service keep track of "Store"
2. The service save the latest Store to disk

3. The service recove from "Store"
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
			grpc.MaxCallRecvMsgSize(s.cfg.GrpcMaxCallRecvMsgSize),
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

	tmplog.Printf("connecting to grpc server: %s", s.cfg.FullNodeServerEndpoint)
	conn, err := grpc.DialContext(s.ctx, s.cfg.FullNodeServerEndpoint, dialOpts...)
	if err != nil {
		panic(err)
	}
	s.conn = conn
	s.lightClientServer = ethpb.NewLightClientClient(s.conn)

	// 1. Recover from a file location if possible
	s.initStore()
	tmplog.Println("Initialized store.")

	// 2. Use hard coded trustedCurrentCommitteeRoot
	s.skipSync()
	tmplog.Println("Finished skip sync.")

	// 3. sync
	go s.sync(s.ctx)
}

func (s *Service) initStore() {
	// Attempt to recover from disk
	s.tryRecoverSnapshot()
	if s.Store != nil {
		return
	}

	// Fetch a skip-sync-update
	key, err := base64.StdEncoding.DecodeString(s.cfg.TrustedCurrentCommitteeRoot)
	if err != nil {
		panic(err)
	}
	update, err := s.lookupSkipSync(key)
	if err != nil {
		panic("cannot initialize")
	}

	// Validate
	err = validateMerkleSkipSyncUpdate(update)
	if err != nil {
		panic(err) // TODO: retry another skip-sync-update
	}
	// Initialize a Store
	store := &ethpb.LightClientStore{
		Updates: []*ethpb.LightClientUpdate{},
		Snapshot: &ethpb.LightClientSnapshot{
			Header:               update.Header,
			CurrentSyncCommittee: update.CurrentSyncCommittee,
			NextSyncCommittee:    update.NextSyncCommittee,
		},
	}
	s.Store = store
}

// Skip sync until we could use normal updates (ethereum.eth.v1alpha1.LightClient.GetUpdates) to sync
func (s *Service) skipSync() {
	resp, err := s.lightClientServer.GetUpdates(s.ctx, &empty.Empty{})
	if err != nil || len(resp.GetUpdates()) < 1 {
		panic(err) // TODO: retry
	}
	update := resp.Updates[0] // TODO: retry
	err = validateMerkleAll(update)
	if err != nil {
		panic(err) // TODO: try other updates, or try another request
	}
	expectedNextSyncComm := update.NextSyncCommittee
	expectedNextSyncCommRoot, err := expectedNextSyncComm.HashTreeRoot()
	if err != nil {
		panic(err)
	}

	storedNextSyncCommRoot, err := s.Store.Snapshot.NextSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	count := 0
	// Skip sync until we could use the expected next-sync-committee
	for {
		// Stop the skip sync once we could use LightClientUpdate to sync normally
		if storedNextSyncCommRoot == expectedNextSyncCommRoot {
			break
		}

		skipSyncKey, err := s.Store.Snapshot.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		skipSyncUpdate, err := s.lookupSkipSync(skipSyncKey[:])
		if err != nil && strings.Contains(err.Error(), "cannot find skip sync error") {
			tmplog.Println(err)
			tmplog.Println("keep trying to find the same skip-sync object; we might be running a infinite loop")
			continue // TODO: This could go on infinite loop or a really long loop
		} else if err != nil {
			panic(err)
		}
		err = validateMerkleSkipSyncUpdate(skipSyncUpdate)
		if err != nil {
			panic(err)
		}

		// Skip sync
		s.simpleProcessSkipSyncUpdate(skipSyncUpdate) // TODO: there might be other checks?

		storedCurrentSyncCommRoot, err := s.Store.Snapshot.CurrentSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		storedNextSyncCommRoot, err := s.Store.Snapshot.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}

		skipSyncNextSyncCommRoot, err := skipSyncUpdate.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		skipSyncCurrentSyncCommRoot, err := skipSyncUpdate.CurrentSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}

		tmplog.Printf("-----skip sync: %d------", count)
		tmplog.Println("expected: ", update.Header)
		tmplog.Println("Store:    ", s.Store.Snapshot.Header)

		tmplog.Println("expected Next:   ", base64.StdEncoding.EncodeToString(expectedNextSyncCommRoot[:]))
		tmplog.Println("stored   Next:   ", base64.StdEncoding.EncodeToString(storedNextSyncCommRoot[:]))
		tmplog.Println("skip     Next:   ", base64.StdEncoding.EncodeToString(skipSyncNextSyncCommRoot[:]))

		tmplog.Println("Store    Current:", base64.StdEncoding.EncodeToString(storedCurrentSyncCommRoot[:]))
		tmplog.Println("skip     Current:", base64.StdEncoding.EncodeToString(skipSyncCurrentSyncCommRoot[:]))
		tmplog.Println("skip     Period:", int(skipSyncUpdate.Header.Slot)/int(EpochsPerSyncCommitteePeriod))

		time.Sleep(time.Second * 1) // TODO: debug only
		count += 1
	}
}

func (s *Service) sync(ctx context.Context) {
	count := 0
	for {
		currentSyncRoot, _ := s.Store.Snapshot.CurrentSyncCommittee.HashTreeRoot()
		nextSyncRoot, _ := s.Store.Snapshot.NextSyncCommittee.HashTreeRoot()
		resp, err := s.lightClientServer.GetUpdates(s.ctx, &emptypb.Empty{})
		if err != nil {
			panic(err) // TODO: retry instead
		}
		update := resp.Updates[0]
		updateNextSyncRoot, _ := update.NextSyncCommittee.HashTreeRoot()

		tmplog.Printf("----------%d----------", count)
		tmplog.Println("header slot         :", s.Store.Snapshot.Header.Slot)
		tmplog.Println("header period       :", int(s.Store.Snapshot.Header.Slot)/int(EpochsPerSyncCommitteePeriod))
		tmplog.Println("header current sync :", base64.StdEncoding.EncodeToString(currentSyncRoot[:]))
		tmplog.Println("header next sync    :", base64.StdEncoding.EncodeToString(nextSyncRoot[:]))
		tmplog.Println("update slot         :", update.Header.Slot)
		tmplog.Println("update period       :", int(update.Header.Slot)/int(EpochsPerSyncCommitteePeriod))
		tmplog.Println("update next sync    :", base64.StdEncoding.EncodeToString(updateNextSyncRoot[:]))

		s.processLightClientUpdate(update)

		s.saveSnapshot()
		time.Sleep(time.Second * 12) // TODO: use a slot tick instead
		count += 1
	}
	tmplog.Println("xxx light client sync done processing xxx")
}

func (s *Service) simpleProcessSkipSyncUpdate(update *ethpb.SkipSyncUpdate) {
	s.Store.Snapshot.Header = update.FinalityHeader
	s.Store.Snapshot.CurrentSyncCommittee = update.CurrentSyncCommittee
	s.Store.Snapshot.NextSyncCommittee = update.NextSyncCommittee
}

func (s *Service) processLightClientUpdate(update *ethpb.LightClientUpdate) {
	processLightClientUpdate(s.Store, update, s.Store.Snapshot.Header.Slot, GenesisValidatorsRoot)
}

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

func (s *Service) saveSnapshot() {
	snapshotBytes, err := proto.Marshal(s.Store)
	if err != nil {
		panic(err)
	}
	filename := s.getStoreFileName()
	tmplog.Printf("Saving store to: %s", filename)
	err = os.WriteFile(filename, snapshotBytes, 0666)
	if err != nil {
		panic(err)
	}
}

func (s *Service) getStoreFileName() string {
	dir, err := filepath.Abs(s.cfg.DataDir)
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, "store.proto-byte")
}

func (s *Service) tryRecoverSnapshot() {
	filename := s.getStoreFileName()
	tmplog.Printf("Recovering store from: %s", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		tmplog.Println(err)
		return
	}
	store := &ethpb.LightClientStore{}
	err = proto.Unmarshal(data, store)
	if err != nil {
		panic(err)
	}
	s.Store = store
	tmplog.Println("Recovered store.snapshot: %s", SnapshotToString(s.Store.Snapshot))
}

func SnapshotToString(snapshot *ethpb.LightClientSnapshot) string {
	_currentSyncCommRoot, err := snapshot.CurrentSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	currentSyncCommRoot := base64.StdEncoding.EncodeToString(_currentSyncCommRoot[:])

	_nextSyncCommRoot, err := snapshot.NextSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	nextSyncCommRoot := base64.StdEncoding.EncodeToString(_nextSyncCommRoot[:])
	headerStr := snapshot.Header.String()
	s := fmt.Sprintf("header=%s|current-sync=%s|next-sync=%s", headerStr, currentSyncCommRoot, nextSyncCommRoot)

	return s
}

// TODO: hacky
func (s *Service) lookupSkipSync(key []byte) (*ethpb.SkipSyncUpdate, error) {
	update, err := s.lightClientServer.GetSkipSyncUpdate(s.ctx, &ethpb.SkipSyncRequest{
		Key: key,
	})
	if err != nil && strings.Contains(err.Error(), "cannot find skip sync error") {
		recommendedKeyStr := lightclientdebug.GetTrustedCurrentCommitteeRoot()
		keyStr := base64.StdEncoding.EncodeToString(key)

		errMsg := fmt.Sprintf("cannot find skip sync error; requesting key=%s", keyStr)
		tmplog.Printf(errMsg)
		tmplog.Printf("The backing beacon-node, %s, does not have the requested ski-sync-update", s.cfg.FullNodeServerEndpoint)
		tmplog.Printf("Consider starting the lightnode with: --trusted-current-committee-root='%s'", recommendedKeyStr)
		return nil, fmt.Errorf(errMsg)
	} else if err != nil {
		tmplog.Printf("check if address %s is available", s.cfg.FullNodeServerEndpoint)
		panic(err) // TODO: Consider retry
	}

	return update, nil
}
