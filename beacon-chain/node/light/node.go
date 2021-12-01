package node

import (
	"context"
	lightrpc "github.com/prysmaticlabs/prysm/beacon-chain/node/light/rpc"
	lightsync "github.com/prysmaticlabs/prysm/beacon-chain/node/light/sync"
	"github.com/prysmaticlabs/prysm/cmd"
	"github.com/prysmaticlabs/prysm/cmd/beacon-chain/flags"
	"github.com/prysmaticlabs/prysm/runtime"
	"github.com/prysmaticlabs/prysm/runtime/debug"
	tmplog "log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
)

type LightNode struct {
	cliCtx   *cli.Context
	ctx      context.Context
	cancel   context.CancelFunc
	services *runtime.ServiceRegistry
	lock     sync.RWMutex
	stop     chan struct{} // Channel to wait for termination notifications.
	//finalizedStateAtStartUp state.BeaconState
	//serviceFlagOpts         *serviceFlagOpts
}

func New(cliCtx *cli.Context) (*LightNode, error) {
	registry := runtime.NewServiceRegistry()

	ctx, cancel := context.WithCancel(cliCtx.Context)
	beacon := &LightNode{
		cliCtx:   cliCtx,
		ctx:      ctx,
		cancel:   cancel,
		services: registry,
		stop:     make(chan struct{}),
		//stateFeed:               new(event.Feed),
		//blockFeed:               new(event.Feed),
		//opFeed:                  new(event.Feed),
		//attestationPool:         attestations.NewPool(),
		//exitPool:                voluntaryexits.NewPool(),
		//slashingsPool:           slashings.NewPool(),
		//syncCommitteePool:       synccommittee.NewPool(),
		//slasherBlockHeadersFeed: new(event.Feed),
		//slasherAttestationsFeed: new(event.Feed),
		//serviceFlagOpts: &serviceFlagOpts{},
	}

	//if err := beacon.startDB(cliCtx, depositAddress); err != nil {
	//	return nil, err
	//}

	if err := beacon.registerSyncService(); err != nil {
		return nil, err
	}

	if err := beacon.registerRpcService(); err != nil {
		return nil, err
	}

	//if err := beacon.registerGRPCGateway(); err != nil {
	//	return nil, err
	//}

	return beacon, nil
}

func (b *LightNode) registerSyncService() error {
	//var chainService *blockchain.Service
	//if err := b.services.FetchService(&chainService); err != nil {
	//	return err
	//}
	//var syncService *initialsync.Service
	//if err := b.services.FetchService(&syncService); err != nil {
	//	return err
	//}

	svc := &lightsync.Service{}

	return b.services.RegisterService(svc)
}

func (b *LightNode) registerRpcService() error {
	host := b.cliCtx.String(flags.RPCHostLight.Name)
	port := b.cliCtx.String(flags.RPCPortLight.Name)
	//beaconMonitoringHost := b.cliCtx.String(cmd.MonitoringHostFlag.Name)
	//beaconMonitoringPort := b.cliCtx.Int(flags.MonitoringPortFlag.Name)
	//cert := b.cliCtx.String(flags.CertFlag.Name)
	//key := b.cliCtx.String(flags.KeyFlag.Name)
	//mockEth1DataVotes := b.cliCtx.Bool(flags.InteropMockEth1DataVotesFlag.Name)

	maxMsgSize := b.cliCtx.Int(cmd.GrpcMaxCallRecvMsgSizeFlag.Name)
	//enableDebugRPCEndpoints := b.cliCtx.Bool(flags.EnableDebugRPCEndpoints.Name)

	tmplog.Println(host)
	tmplog.Println(port)
	tmplog.Println(maxMsgSize)

	svc, err := lightrpc.NewService(b.ctx, &lightrpc.Config{
		Host:       host,
		Port:       port,
		MaxMsgSize: maxMsgSize,
	})
	if err != nil {
		panic(err)
	}

	return b.services.RegisterService(svc)
}

// Start the BeaconNode and kicks off every registered service.
func (b *LightNode) Start() {
	b.lock.Lock()
	b.services.StartAll()

	stop := b.stop
	b.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		debug.Exit(b.cliCtx) // Ensure trace and CPU profile data are flushed.
		go b.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.WithField("times", i-1).Info("Already shutting down, interrupt more to panic")
			}
		}
		panic("Panic closing the beacon node")
	}()

	// Wait for stop channel to be closed.
	<-stop
}

// Close handles graceful shutdown of the system.
func (b *LightNode) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	log.Info("Stopping beacon node")
	b.services.StopAll()
	//if err := b.db.Close(); err != nil {
	//	log.Errorf("Failed to close database: %v", err)
	//}
	//b.collector.unregister()
	b.cancel()
	close(b.stop)
}
