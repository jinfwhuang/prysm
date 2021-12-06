package node

import (
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/light-update"
	initialsync "github.com/prysmaticlabs/prysm/beacon-chain/sync/initial-sync"
)

// TODO: hack; using a separate file only to make merging easier.

func (b *BeaconNode) registerLightUpdateService() error {
	var chainService *blockchain.Service
	if err := b.services.FetchService(&chainService); err != nil {
		return err
	}
	var syncService *initialsync.Service
	if err := b.services.FetchService(&syncService); err != nil {
		return err
	}
	svc := light.New(b.ctx, &light.Config{
		Database:                    b.db,
		StateGen:                    b.stateGen,
		HeadFetcher:                 chainService,
		FinalizationFetcher:         chainService,
		StateNotifier:               b,
		TimeFetcher:                 chainService,
		SyncChecker:                 syncService,
		LightClientUpdatesQueueSize: 32, // TODO: jin; allowed this to be passed in from command line
	})

	return b.services.RegisterService(svc)
}

func (b *BeaconNode) fetchLightUpdateService() *light.Service {
	var s *light.Service
	if err := b.services.FetchService(&s); err != nil {
		panic(err)
	}
	return s
}
