package blockchain

import (
	"context"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	tmplog "log"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

// Use the blocks to build light-client-updates
func (s *Service) buildLightClientUpdates(ctx context.Context, block block.SignedBeaconBlock) error {
	//blockCopy := block.Copy()

	tmplog.Println(block)
	tmplog.Println("------------------")
	tmplog.Println("------------------")
	tmplog.Println("len", s.queue.GetLen())
	tmplog.Println("cap", s.queue.GetCap())

	// Use the block to build a light-client-update queue

	return nil
}
