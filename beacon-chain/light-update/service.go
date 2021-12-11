package light

import (
	"context"
	"encoding/base64"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/iface"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	syncSrv "github.com/prysmaticlabs/prysm/beacon-chain/sync"
	"github.com/prysmaticlabs/prysm/cache/fifo"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	block "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
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
	//_blk, st, err := s.GetChainHeadAndState(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//blk := _blk.Block()
	//if err != nil {
	//	panic(err)
	//}
	//root, err := blk.HashTreeRoot()

	root, err := s.cfg.HeadFetcher.HeadRoot(ctx)
	if err != nil {
		panic(err)
	}

	////ckp := st.FinalizedCheckpoint()
	//s.learnState(ctx, root[:])
	//
	err = s.processFinalizedEvent(ctx, root)
	if err != nil {
		panic(err)
	}
}

func printBlk(blk block.SignedBeaconBlock) {
	_header, err := blk.Header()
	if err != nil {
		panic(err)
	}
	header := _header.Header

	blkRoot, err := header.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	tmplog.Println("-------block-----------")
	tmplog.Println(header)
	tmplog.Println("blk root", blkRoot)
	tmplog.Println("----------------")
}

func printState(st state.BeaconState) {
	//blkRoot, err := state.LatestBlockHeader().HashTreeRoot()
	//if err != nil {
	//	panic(err)
	//}
	stRoot, err := st.HashTreeRoot(context.Background())
	if err != nil {
		panic(err)
	}

	tmplog.Println("-------state-----------")
	tmplog.Println("state root", stRoot)
	tmplog.Println("state checkpoint", st.FinalizedCheckpoint())
	tmplog.Println("----------------")
	//tmplog.Println("blk root", blkRoot)
	//tmplog.Println("blk header", state.LatestBlockHeader())

}

func (s *Service) learnState(ctx context.Context, root []byte) {
	tmplog.Println("-------learning-----------")
	tmplog.Println("root", base64.StdEncoding.EncodeToString(root))
	blk, err := s.getBlock(ctx, root)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	//header, err := blk.Header()
	//if err != nil {
	//	tmplog.Println(err)
	//	log.Error(err)
	//	return
	//}
	blkRoot, err := blk.Block().HashTreeRoot()
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("block root", base64.StdEncoding.EncodeToString(blkRoot[:]))

	st, err := s.getState(ctx, root)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	stRoot, err := st.HashTreeRoot(context.Background())
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("state root", base64.StdEncoding.EncodeToString(stRoot[:]))
	tmplog.Println("checkpoint root", base64.StdEncoding.EncodeToString(st.FinalizedCheckpoint().Root))
	tmplog.Println("state checkpoint", st.FinalizedCheckpoint())

	fRoot := st.FinalizedCheckpoint().Root

	fBlk, err := s.getBlock(ctx, fRoot)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("---")
	tmplog.Println("f root", base64.StdEncoding.EncodeToString(fRoot))
	tmplog.Println("f block", fBlk)

	fSt, err := s.getState(ctx, fRoot)
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	fStRoot, err := fSt.HashTreeRoot(context.Background())
	if err != nil {
		tmplog.Println(err)
		log.Error(err)
		return
	}
	tmplog.Println("f state root", base64.StdEncoding.EncodeToString(fStRoot[:]))
	tmplog.Println("----------------")
}

func Equal(a, b [32]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
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
