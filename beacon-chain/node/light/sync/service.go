// Package blockchain defines the life-cycle of the blockchain at the core of
// Ethereum, including processing of new blocks and attestations using proof of stake.
package sync

import (
	"context"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	tmplog "log"
	"time"
)

// headSyncMinEpochsAfterCheckpoint defines how many epochs should elapse after known finalization
// checkpoint for head sync to be triggered.
const headSyncMinEpochsAfterCheckpoint = 128

type Service struct {
	cfg    *config
	ctx    context.Context
	cancel context.CancelFunc
	//genesisTime           time.Time
	//head                  *head
	//headLock              sync.RWMutex
	//genesisRoot           [32]byte
	finalizedCheckpt *ethpb.Checkpoint
	blockHeader      *ethpb.BeaconBlockHeader
	//justifiedBalances     []uint64
	//justifiedBalances *stateBalanceCache
	//wsVerifier        *WeakSubjectivityVerifier
}

// config options for the service.
type config struct {
	//BeaconBlockBuf          int
	//ChainStartFetcher       powchain.ChainStartFetcher
	//BeaconDB                db.HeadAccessDatabase
	//DepositCache            *depositcache.DepositCache
	//AttPool                 attestations.Pool
	//ExitPool                voluntaryexits.PoolManager
	//SlashingPool            slashings.PoolManager
	//P2p                     p2p.Broadcaster
	//MaxRoutines             int
	//StateNotifier           statefeed.Notifier
	//ForkChoiceStore         f.ForkChoicer
	//AttService              *attestations.Service
	//StateGen                *stategen.State
	//SlasherAttestationsFeed *event.Feed
	//WeakSubjectivityCheckpt *ethpb.Checkpoint
	//FinalizedStateAtStartUp state.BeaconState
}

// NewService instantiates a new block service instance that will
// be registered into a running beacon node.
func NewService(ctx context.Context) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	svr := &Service{
		ctx:    ctx,
		cancel: cancel,
		cfg:    &config{},
	}
	return svr, nil
}

// Start a blockchain service's main event loop.
func (s *Service) Start() {
	go s.process(s.ctx)
}

// processChainStartTime initializes a series of deposits from the ChainStart deposits in the eth1
// deposit contract, initializes the beacon chain's state, and kicks off the beacon chain.
func (s *Service) process(ctx context.Context) {
	count := 0
	for {
		tmplog.Println("processing", count)
		time.Sleep(time.Second * 10)
		count += 1
	}
	tmplog.Println("xxx light client sync done processing xxx")
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
