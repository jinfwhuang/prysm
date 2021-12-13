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
	eth2_types "github.com/prysmaticlabs/eth2-types"
	grpcutil "github.com/prysmaticlabs/prysm/api/grpc"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/time/slots"
	"os"
	"path/filepath"
	"strings"

	//lightnode "github.com/prysmaticlabs/prysm/cmd/lightclient/node"
	lightclientdebug "github.com/prysmaticlabs/prysm/cmd/light-client/debug"
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
	EpochsPerSyncCommitteePeriod        = params.BeaconConfig().EpochsPerSyncCommitteePeriod
	UpdateTimeout                       = params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod))
	MinSyncCommitteeParticipants uint64 = 1
)

func (s SyncMode) String() string {
	return []string{
		"latest",
		"finality",
	}[s]
}

type Service struct {
	cfg         *Config
	ctx         context.Context
	cancel      context.CancelFunc
	lock        sync.RWMutex
	genesisTime time.Time

	dataDir string
	Store   *ethpb.LightClientStore

	// grpc
	conn              *grpc.ClientConn
	lightUpdateClient ethpb.LightUpdateClient
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
		ctx:         ctx,
		cancel:      cancel,
		cfg:         cfg,
		genesisTime: time.Unix(1606824023, 0), // 2020-12-01 04:00:23 -0800 PST
		// nano: 32536808316530000

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
	s.lightUpdateClient = ethpb.NewLightUpdateClient(s.conn)

	// TODO: re-enable
	// 1. Recover from a file location if possible
	s.initStore()
	tmplog.Println("Initialized store.")

	// 2. Sky sync to current sync-committee
	s.skipSync()
	tmplog.Println("Finished skip sync.")

	// 3. sync
	go s.sync(s.ctx)
}

func (s *Service) initStore() {
	// TODO: re-enable
	//// Attempt to recover from disk
	//s.tryRecoverStore()
	//if s.Store != nil {
	//	return
	//}

	// Init store for the first time
	// Fetch a skip-sync-update
	tmplog.Println("Initializing a LightClientStore ...")
	key, err := base64.StdEncoding.DecodeString(s.cfg.TrustedCurrentCommitteeRoot)
	if err != nil {
		panic(err)
	}
	update, err := s.lookupSkipSync(key)
	if err != nil {
		panic(err)
	}

	// Validate
	err = validateMerkleSkipSyncUpdate(update)
	if err != nil {
		panic(err) // TODO: retry another skip-sync-update
	}

	// Initialize a Store
	store := &ethpb.LightClientStore{
		FinalizedHeader:               update.FinalityHeader,
		OptimisticHeader:              update.AttestedHeader,
		CurrentSyncCommittee:          update.CurrentSyncCommittee,
		NextSyncCommittee:             update.NextSyncCommittee,
		PreviousMaxActiveParticipants: update.SyncCommitteeBits.Count(),
		CurrentMaxActiveParticipants:  update.SyncCommitteeBits.Count(),
		BestValidUpdate:               toLightClientUpdate(update),
	}

	tmplog.Println("-----")
	tmplog.Println("store.OptimisticHeader", store.OptimisticHeader == nil)
	tmplog.Println("store.BestValidUpdate", store.BestValidUpdate == nil)
	tmplog.Println("store.CurrentSyncCommittee", store.CurrentSyncCommittee == nil)
	tmplog.Println("store.NextSyncCommittee", store.NextSyncCommittee == nil)
	tmplog.Println("-----")

	s.Store = store
}

// Skip sync until we could use normal updates (ethereum.eth.v1alpha1.LightClient.GetUpdates) to sync
// Skip sync only use update.finality_header
func (s *Service) skipSync() {
	resp, err := s.lightUpdateClient.GetUpdates(s.ctx, &empty.Empty{})
	if err != nil || len(resp.GetUpdates()) < 1 {
		panic(err) // TODO: implement retry retry
	}
	update := resp.Updates[0] // TODO: we could use other updates from resp
	err = validateMerkleLightClientUpdate(update)
	if err != nil {
		panic(err) // TODO: implement retry
	}
	expectedNextSyncComm := update.NextSyncCommittee
	expectedNextSyncCommRoot, err := expectedNextSyncComm.HashTreeRoot()
	if err != nil {
		panic(err) // TODO: implement retry
	}

	storedNextSyncCommRoot, err := s.Store.NextSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}

	// Skip sync until we could use the expected next-sync-committee
	count := 0
	for {
		// Stop the skip sync once we could use LightClientUpdate to sync normally
		if storedNextSyncCommRoot == expectedNextSyncCommRoot {
			tmplog.Println("----")
			tmplog.Println("Successfully skip-sync")
			tmplog.Printf(StoreToString(s.Store))
			tmplog.Println("----")
			break
		}

		skipSyncKey, err := s.Store.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		skipSyncUpdate, err := s.lookupSkipSync(skipSyncKey[:])
		if err != nil && strings.Contains(err.Error(), "cannot find skip sync error") {
			tmplog.Println(err)
			tmplog.Println("keep trying to find the same skip-sync object; we might be running a infinite loop")
			time.Sleep(time.Second * 1) // TODO: debug only
			count += 1
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

		// DEBUG messages only
		skipSyncCurrentSyncCommRoot, err := skipSyncUpdate.CurrentSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		skipSyncNextSyncCommRoot, err := skipSyncUpdate.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		storedCurrentSyncCommRoot, err := s.Store.CurrentSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		storedNextSyncCommRoot, err = s.Store.NextSyncCommittee.HashTreeRoot()
		if err != nil {
			panic(err)
		}
		tmplog.Printf("-----skip sync: %d------", count)
		tmplog.Println("expected: ", update.FinalityHeader)
		tmplog.Println("Store:    ", s.Store.FinalizedHeader)

		tmplog.Println("Store    Current:", base64.StdEncoding.EncodeToString(storedCurrentSyncCommRoot[:]))
		tmplog.Println("skip     Current:", base64.StdEncoding.EncodeToString(skipSyncCurrentSyncCommRoot[:]))

		tmplog.Println("expected Next:   ", base64.StdEncoding.EncodeToString(expectedNextSyncCommRoot[:]))
		tmplog.Println("skip     Next:   ", base64.StdEncoding.EncodeToString(skipSyncNextSyncCommRoot[:]))
		tmplog.Println("stored   Next:   ", base64.StdEncoding.EncodeToString(storedNextSyncCommRoot[:]))

		time.Sleep(time.Second * 1) // TODO: debug only
		count += 1
	}
}

func (s *Service) sync(ctx context.Context) {
	count := 0
	for {
		currentSyncRoot, _ := s.Store.CurrentSyncCommittee.HashTreeRoot()
		nextSyncRoot, _ := s.Store.NextSyncCommittee.HashTreeRoot()
		resp, err := s.lightUpdateClient.GetUpdates(s.ctx, &emptypb.Empty{})
		if err != nil {
			panic(err) // TODO: retry instead
		}
		update := resp.Updates[0]

		// DEBUG
		updateNextSyncRoot, _ := update.NextSyncCommittee.HashTreeRoot()
		tmplog.Printf("----------%d----------", count)
		tmplog.Println("optic slot          : ", s.Store.OptimisticHeader.Slot)
		tmplog.Println("optic period        :", int(s.Store.OptimisticHeader.Slot)/int(EpochsPerSyncCommitteePeriod))
		tmplog.Println("finality slot       : ", s.Store.FinalizedHeader.Slot)
		tmplog.Println("finality period     :", int(s.Store.FinalizedHeader.Slot)/int(EpochsPerSyncCommitteePeriod))
		tmplog.Println("current sync        :", base64.StdEncoding.EncodeToString(currentSyncRoot[:]))
		tmplog.Println("next sync           :", base64.StdEncoding.EncodeToString(nextSyncRoot[:]))
		tmplog.Printf("--")
		tmplog.Println("update slot         :", update.AttestedHeader.Slot)
		tmplog.Println("update period       :", int(update.AttestedHeader.Slot)/int(EpochsPerSyncCommitteePeriod))
		tmplog.Println("update next sync    :", base64.StdEncoding.EncodeToString(updateNextSyncRoot[:]))
		// END OF DEBUG

		s.processLightClientUpdate(update)
		s.saveSnapshot()

		time.Sleep(time.Second * 12) // TODO: use a slot tick instead
		count += 1
	}
	tmplog.Println("xxx light client sync done processing xxx")
}

func (s *Service) simpleProcessSkipSyncUpdate(update *ethpb.SkipSyncUpdate) {
	s.Store.FinalizedHeader = update.FinalityHeader
	s.Store.CurrentSyncCommittee = update.CurrentSyncCommittee
	s.Store.NextSyncCommittee = update.NextSyncCommittee
}

func (s *Service) processLightClientUpdate(update *ethpb.LightClientUpdate) {
	currentSlot := s.CurrentSlot()
	processLightClientUpdate(s.Store, update, currentSlot, GenesisValidatorsRoot)
}

func (s *Service) CurrentSlot() eth2_types.Slot {
	return slots.CurrentSlot(uint64(s.genesisTime.Unix()))
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

func (s *Service) tryRecoverStore() {
	filename := s.getStoreFileName()
	tmplog.Printf("Recovering store from: %s", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		tmplog.Println(err)
		tmplog.Println("cannot recover a store")
		return
	}
	store := &ethpb.LightClientStore{}
	err = proto.Unmarshal(data, store)
	if err != nil {
		panic(err)
	}
	s.Store = store
	tmplog.Printf("Recovered store: %s", StoreToString(s.Store))
}

func StoreToString(store *ethpb.LightClientStore) string {
	_currentSyncCommRoot, err := store.CurrentSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	currentSyncCommRoot := base64.StdEncoding.EncodeToString(_currentSyncCommRoot[:])

	_nextSyncCommRoot, err := store.NextSyncCommittee.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	nextSyncCommRoot := base64.StdEncoding.EncodeToString(_nextSyncCommRoot[:])

	s := fmt.Sprintf(
		"optimistic header=%s \n"+
			"finality header=%s \n"+
			"current-sync=%s \n "+
			"next-sync=%s",
		store.OptimisticHeader.String(), store.FinalizedHeader.String(), currentSyncCommRoot, nextSyncCommRoot,
	)

	return s
}

// TODO: hacky
func (s *Service) lookupSkipSync(key []byte) (*ethpb.SkipSyncUpdate, error) {
	update, err := s.lightUpdateClient.GetSkipSyncUpdate(s.ctx, &ethpb.SkipSyncRequest{
		Key: key,
	})
	if err != nil && strings.Contains(err.Error(), "cannot find skip sync error") {
		tmplog.Println("----debug message-----")
		recommendedKeyStr := lightclientdebug.GetTrustedCurrentCommitteeRoot()
		keyStr := base64.StdEncoding.EncodeToString(key)

		errMsg := fmt.Sprintf("cannot find skip sync error; requesting key=%s", keyStr)
		tmplog.Printf(errMsg)
		tmplog.Printf("The backing beacon-node, %s, does not have the requested ski-sync-update", s.cfg.FullNodeServerEndpoint)
		tmplog.Printf("---------------")
		tmplog.Printf("---------------")
		tmplog.Printf("Consider starting the lightnode with:")
		tmplog.Printf("--trusted-current-committee-root='%s'", recommendedKeyStr)
		tmplog.Printf("---------------")
		tmplog.Printf("---------------")
		return nil, fmt.Errorf(errMsg)
	} else if err != nil {
		tmplog.Printf("check if address %s is available", s.cfg.FullNodeServerEndpoint)
		panic(err) // TODO: Consider retry
	}

	return update, nil
}
