package node

import (
	"context"
	"github.com/prysmaticlabs/prysm/cmd"
	"github.com/prysmaticlabs/prysm/cmd/beacon-chain/flags"
	lightutil "github.com/prysmaticlabs/prysm/cmd/lightclient/util"

	//"github.com/prysmaticlabs/prysm/cmd/lightclient"
	lightrpc "github.com/prysmaticlabs/prysm/cmd/lightclient/rpc"
	lightsync "github.com/prysmaticlabs/prysm/cmd/lightclient/sync"
	"github.com/prysmaticlabs/prysm/runtime"
	"github.com/prysmaticlabs/prysm/runtime/debug"
	tmplog "log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

func (ln *LightNode) registerSyncService() error {
	//var chainService *blockchain.Service
	//if err := b.services.FetchService(&chainService); err != nil {
	//	return err
	//}
	//var syncService *initialsync.Service
	//if err := b.services.FetchService(&syncService); err != nil {
	//	return err
	//}

	//grpcRetries := ln.cliCtx.Uint(flags.GrpcRetriesFlag.Name)
	//grpcRetryDelay := ln.cliCtx.Duration(flags.GrpcRetryDelayFlag.Name)
	//maxCallRecvMsgSize := ln.cliCtx.Int(cmd.GrpcMaxCallRecvMsgSizeFlag.Name)
	//dataDir := ln.cliCtx.String(cmd.DataDirFlag.Name)

	// TODO: Link the configuration variables with flags

	svc, err := lightsync.NewService(ln.ctx, &lightsync.Config{
		TrustedCurrentCommitteeRoot: "UeSv92gwGs+DSk34NqOaCM1DaU9zyclQE6Tc9morK0M=",
		GrpcRetryDelay:              time.Second * 5,
		GrpcRetries:                 5,
		MaxCallRecvMsgSize:          1 << 22,
		GrpcEndpoint:                "127.0.0.1:4000",
	})
	if err != nil {
		panic(err)
	}

	return ln.services.RegisterService(svc)
}

func (ln *LightNode) registerRpcService() error {
	host := ln.cliCtx.String(flags.RPCHostLight.Name)
	port := ln.cliCtx.String(flags.RPCPortLight.Name)
	//beaconMonitoringHost := b.cliCtx.String(cmd.MonitoringHostFlag.Name)
	//beaconMonitoringPort := b.cliCtx.Int(flags.MonitoringPortFlag.Name)
	//cert := b.cliCtx.String(flags.CertFlag.Name)
	//key := b.cliCtx.String(flags.KeyFlag.Name)
	//mockEth1DataVotes := b.cliCtx.Bool(flags.InteropMockEth1DataVotesFlag.Name)

	maxMsgSize := ln.cliCtx.Int(cmd.GrpcMaxCallRecvMsgSizeFlag.Name)
	//enableDebugRPCEndpoints := b.cliCtx.Bool(flags.EnableDebugRPCEndpoints.Name)

	tmplog.Println(host)
	tmplog.Println(port)
	tmplog.Println(maxMsgSize)

	var lightsyncService *lightsync.Service
	if err := ln.services.FetchService(&lightsyncService); err != nil {
		return err
	}

	svc, err := lightrpc.NewService(ln.ctx, &lightrpc.Config{
		Host:        host,
		Port:        port,
		MaxMsgSize:  maxMsgSize,
		SyncService: lightsyncService,
	})
	if err != nil {
		panic(err)
	}

	return ln.services.RegisterService(svc)
}

// Start the BeaconNode and kicks off every registered service.
func (ln *LightNode) Start() {
	ln.lock.Lock()
	ln.services.StartAll()

	stop := ln.stop
	ln.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		lightutil.Log.Info("Got interrupt, shutting down...")
		debug.Exit(ln.cliCtx) // Ensure trace and CPU profile data are flushed.
		go ln.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				lightutil.Log.WithField("times", i-1).Info("Already shutting down, interrupt more to panic")
			}
		}
		panic("Panic closing the beacon node")
	}()

	// Wait for stop channel to be closed.
	<-stop
}

// Close handles graceful shutdown of the system.
func (ln *LightNode) Close() {
	ln.lock.Lock()
	defer ln.lock.Unlock()

	lightutil.Log.Info("Stopping beacon node")
	ln.services.StopAll()
	//if err := b.db.Close(); err != nil {
	//	log.Errorf("Failed to close database: %v", err)
	//}
	//b.collector.unregister()
	ln.cancel()
	close(ln.stop)
}
