package light

import (
	"context"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/iface"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	syncSrv "github.com/prysmaticlabs/prysm/beacon-chain/sync"
	"github.com/prysmaticlabs/prysm/cache/fifo"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	block "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	"sync"
	"time"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile) // TODO: hack
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
	return s.cfg.Database.GetSkipSyncUpdate(ctx, key) // TODO: Is there a better way to cache the data?
}

func (s *Service) GetCurrentSyncComm(ctx context.Context) (*ethpb.SyncCommittee, error) {
	st, err := s.GetHeadState(ctx)
	if err != nil {
		return nil, err
	}
	return st.CurrentSyncCommittee()
}

func (s *Service) run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	s.waitForChainInitialization(ctx)
	log.Info("Chain is initialized")

	// Initialize the service.
	log.Info("Initializing light-update service")
	s.initializeFromHead(ctx)
	log.Info("Initialized light-update service")

	log.Info("Start listening for events that will update light client related queue and db")
	go s.subscribeEvents(ctx)
}

func (s *Service) finalizedBlockOrGenesis(ctx context.Context, cpt *ethpb.Checkpoint) (block.SignedBeaconBlock, error) {
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

func (s *Service) initializeFromHead(ctx context.Context) {
	root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
	if err != nil {
		panic(err)
	}
	err = s.processFinalizedEvent(ctx, root)
	if err != nil {
		panic(err)
	}
}

func (s *Service) waitForChainInitialization(ctx context.Context) {
	stateChannel := make(chan *feed.Event, 1)
	stateSub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()
	defer close(stateChannel)
	for {
		select {
		case stateEvent := <-stateChannel:
			// Wait for us to receive the genesis time via a chain started notification.
			if stateEvent.Type == statefeed.Initialized {
				// Alternatively, if the chain has already started, we then read the genesis
				// time value from this data.
				data, ok := stateEvent.Data.(*statefeed.InitializedData)
				if !ok {
					log.Error(
						"Could not receive chain start notification, want *statefeed.ChainStartedData",
					)
					return
				}
				//s.genesisTime = data.StartTime
				log.WithField("genesisTime", data.StartTime).Info(
					"Received chain initialization event",
				)
				tmplog.Println(data.StartTime)
				tmplog.Println(data.StartTime.Unix())
				tmplog.Println(data.StartTime.UnixMilli())
				tmplog.Println(data.StartTime.UnixMicro())
				tmplog.Println(data.StartTime.UnixNano())
				tmplog.Println(data.StartTime.Nanosecond())
				tmplog.Println(time.UnixMilli(data.StartTime.Unix()))
				return
			}
		case err := <-stateSub.Err():
			log.WithError(err).Error(
				"Could not subscribe to state events",
			)
			return
		case <-ctx.Done():
			return
		}
	}
}
