package blockchain

import (
	"context"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
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

	header, err := block.Header()
	if err != nil {
		return err
	}

	blockRoot := header.Header.BodyRoot
	//stateRoot := block.Block().StateRoot(ctx,)

	state, err := s.cfg.BeaconDB.State(ctx, bytesutil.ToBytes32(blockRoot)) // why bodyRoot instead of stateRoot
	tmplog.Println(state)

	//currentCom := state.CurrentSyncCommittee()
	nextCom, err := state.NextSyncCommittee()
	if err != nil {
		return err
	}
	tmplog.Println(nextCom)

	//nextSyncCommittee, err := block.Block().Body().SyncAggregate()

	update := ethpb.LightClientUpdate{
		Header:            header.Header,
		NextSyncCommittee: nextCom,

		//Header                  *BeaconBlockHeader                                `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
		//NextSyncCommittee       *SyncCommittee                                    `protobuf:"bytes,2,opt,name=next_sync_committee,json=nextSyncCommittee,proto3" json:"next_sync_committee,omitempty"`
		//NextSyncCommitteeBranch [][]byte                                          `protobuf:"bytes,3,rep,name=next_sync_committee_branch,json=nextSyncCommitteeBranch,proto3" json:"next_sync_committee_branch,omitempty" ssz-size:"5,32"`
		//FinalityHeader          *BeaconBlockHeader                                `protobuf:"bytes,4,opt,name=finality_header,json=finalityHeader,proto3" json:"finality_header,omitempty"`
		//FinalityBranch          [][]byte                                          `protobuf:"bytes,5,rep,name=finality_branch,json=finalityBranch,proto3" json:"finality_branch,omitempty" ssz-size:"6,32"`
		//SyncCommitteeBits       github_com_prysmaticlabs_go_bitfield.Bitvector512 `protobuf:"bytes,6,opt,name=sync_committee_bits,json=syncCommitteeBits,proto3" json:"sync_committee_bits,omitempty" cast-type:"github.com/prysmaticlabs/go-bitfield.Bitvector512" ssz-size:"64"`
		//SyncCommitteeSignature  []byte                                            `protobuf:"bytes,7,opt,name=sync_committee_signature,json=syncCommitteeSignature,proto3" json:"sync_committee_signature,omitempty" ssz-size:"96"`
		//ForkVersion             []byte                                            `protobuf:"bytes,8,opt,name=fork_version,json=forkVersion,proto3" json:"fork_version,omitempty" ssz-size:"4"`

	}

	// build LightClientUpdate
	tmplog.Println(update)

	// Build SkipSyncUpdate

	// Use the block to build a light-client-update queue

	return nil
}
