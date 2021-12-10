package light

import (
	"context"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/iface"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	syncSrv "github.com/prysmaticlabs/prysm/beacon-chain/sync"
	"github.com/prysmaticlabs/prysm/cache/fifo"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	block2 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	"sync"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

type Config struct {
	StateGen                    stategen.StateManager
	Database                    iface.Database
	HeadFetcher                 blockchain.HeadFetcher
	FinalizationFetcher         blockchain.FinalizationFetcher
	StateNotifier               statefeed.Notifier
	TimeFetcher                 blockchain.TimeFetcher
	SyncChecker                 syncSrv.Checker
	LightClientUpdatesQueueSize int
}

type Service struct {
	cfg        *Config
	cancelFunc context.CancelFunc
	lock       sync.RWMutex
	Queue      fifo.Queue
}

func New(ctx context.Context, cfg *Config) *Service {
	queue := fifo.NewFixedFifo(cfg.LightClientUpdatesQueueSize) // Light client update Queue

	return &Service{
		cfg:   cfg,
		Queue: &queue,
	}
}

func (s *Service) Start() {
	go s.run()
}

func (s *Service) Stop() error {
	s.cancelFunc()
	return nil
}

func (s *Service) Status() error {
	return nil
}

func (s *Service) GetSkipSyncUpdate(ctx context.Context, key [32]byte) (*ethpb.SkipSyncUpdate, error) {
	return s.cfg.Database.GetSkipSyncUpdate(ctx, key)
}

func (s *Service) GetCurrentSyncComm(ctx context.Context) (*ethpb.SyncCommittee, error) {
	_, state, err := s.GetChainHeadAndState(ctx)
	if err != nil {
		return nil, err
	}
	return state.CurrentSyncCommittee()
}

func (s *Service) run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Initialize the service.
	log.Info("Initializing light-update service")
	s.initializeFromHead(ctx)

	log.Info("Start listening for events that will update light client related queue and db")
	go s.subscribeEvents(ctx)
}

func (s *Service) finalizedBlockOrGenesis(ctx context.Context, cpt *ethpb.Checkpoint) (block2.SignedBeaconBlock, error) {
	checkpointRoot := bytesutil.ToBytes32(cpt.Root)
	block, err := s.cfg.Database.Block(ctx, checkpointRoot)
	if err != nil {
		return nil, err
	}
	if block == nil || block.IsNil() {
		return s.cfg.Database.GenesisBlock(ctx)
	}
	return block, nil
}

func (s *Service) finalizedStateOrGenesis(ctx context.Context, cpt *ethpb.Checkpoint) (state.BeaconState, error) {
	checkpointRoot := bytesutil.ToBytes32(cpt.Root)
	st, err := s.cfg.StateGen.StateByRoot(ctx, checkpointRoot)
	if err != nil {
		return nil, err
	}
	if st == nil || st.IsNil() {
		return s.cfg.Database.GenesisState(ctx)
	}
	return st, nil
}

func (s *Service) initializeFromHead(ctx context.Context) {
	head, beaconState, err := s.GetChainHeadAndState(ctx)
	if err != nil {
		panic(err)
	}

	err = s.processHeadEvent(ctx, head, beaconState)
	if err != nil {
		tmplog.Println(err)
	}

}
