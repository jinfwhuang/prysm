package light

import (
	"context"
	"encoding/hex"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	//vv1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	statev2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	log "github.com/sirupsen/logrus"
	tmplog "log"
	"os"
	//"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	//"github.com/prysmaticlabs/prysm/encoding/ssz"
	//"github.com/prysmaticlabs/prysm/network/forks"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	//"github.com/prysmaticlabs/prysm/time/slots"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype/utils"
)

const (
	finalizedCheckpointStateIndex = 20
	nextSyncCommitteeStateIndex   = 23
)

func (s *Service) subscribeHeadEvent(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.NewHead {
				head, beaconState, err := s.getChainHeadAndState(ctx)
				if err != nil {
					log.Error(err)
					continue
				}

				// TODO: xxx building client updates
				if err := s.buildLightClientUpdates(ctx, head, beaconState); err != nil {
					log.Error(err)
					continue
				}

			}
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) subscribeFinalizedEvent(ctx context.Context) {
	stateChan := make(chan *feed.Event, 1)
	sub := s.cfg.StateNotifier.StateFeed().Subscribe(stateChan)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-stateChan:
			if ev.Type == statefeed.FinalizedCheckpoint {
				blk, beaconState, err := s.parseFinalizedEvent(ctx, ev.Data)
				if err != nil {
					log.Error(err)
					continue
				}
				//tmplog.Println(blk, beaconState)

				s.onFinalizedCheckpoint(ctx, blk, beaconState)

				//if err := s.onFinalized(ctx, blk, beaconState); err != nil {
				//	log.Error(err)
				//	continue
				//}
			}
			//} else {
			//	tmplog.Println(ev.Type)
			//}
		}
	}
}

func (s *Service) parseFinalizedEvent(
	ctx context.Context, eventData interface{},
) (block.SignedBeaconBlock, state.BeaconState, error) {
	finalizedCheckpoint, ok := eventData.(*v1.EventFinalizedCheckpoint)
	if !ok {
		return nil, nil, errors.New("expected finalized checkpoint event")
	}
	checkpointRoot := bytesutil.ToBytes32(finalizedCheckpoint.Block)
	blk, err := s.cfg.Database.Block(ctx, checkpointRoot)
	if err != nil {
		return nil, nil, err
	}
	if blk == nil || blk.IsNil() {
		return nil, nil, err
	}
	st, err := s.cfg.StateGen.StateByRoot(ctx, checkpointRoot)
	if err != nil {
		return nil, nil, err
	}
	if st == nil || st.IsNil() {
		return nil, nil, err
	}
	return blk, st, nil
}

func (s *Service) getChainHeadAndState(ctx context.Context) (block.SignedBeaconBlock, state.BeaconState, error) {
	head, err := s.cfg.HeadFetcher.HeadBlock(ctx)
	if err != nil {
		return nil, nil, err
	}
	if head == nil || head.IsNil() {
		return nil, nil, errors.New("head block is nil")
	}
	st, err := s.cfg.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, nil, errors.New("head state is nil")
	}
	if st == nil || st.IsNil() {
		return nil, nil, err
	}
	return head, st, nil
}

func writetofile(bytearray []byte, root [32]byte) {
	filename := "./tmp/ssz/statev1/" + utils.HashBytes(bytearray) + ".ssz"
	err := os.WriteFile(filename, bytearray, 0666)
	if err != nil {
		panic(err)
	}

	contentHash := utils.HashBytes(bytearray)

	tmplog.Println("----------------------------")
	tmplog.Println("wrote a ssz file:", filename)
	tmplog.Println("content hash:", contentHash)
	tmplog.Println("root hash", hex.EncodeToString(root[:]))
	tmplog.Println("----------------------------")
}

func saveSsz(ctx context.Context, state state.BeaconStateAltair) {
	// Use beaconstatev2.BeaconState
	bytearray, err := state.MarshalSSZ()
	if err != nil {
		panic(err)
	}
	stateRoot, err := state.HashTreeRoot(ctx)
	if err != nil {
		panic(err)
	}
	tmplog.Println("statev2")
	writetofile(bytearray, stateRoot)

	// Use ethpb.BeaconStateAltair
	ethpbState := state.(*statev2.BeaconState).GetState()
	bytearray, err = ethpbState.MarshalSSZ()
	if err != nil {
		panic(err)
	}
	ethpbStateRoot, err := ethpbState.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	tmplog.Println("ethpb BeaconState")
	writetofile(bytearray, ethpbStateRoot)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func verify_serder_same_root(ctx context.Context, state1 state.BeaconState) {
	root1, err := state1.HashTreeRoot(ctx)
	stateStructv1 := state1.(*statev2.BeaconState)
	panicErr(err)

	bytearray, err := state1.MarshalSSZ()
	state2 := mkBeaconStateAltair(bytearray) // From ser-der process
	stateStructv2 := state2.(*statev2.BeaconState)
	root2, err := state2.HashTreeRoot(ctx)

	tmplog.Println(hex.EncodeToString(root1[:]))
	tmplog.Println(hex.EncodeToString(root2[:]))

	tmplog.Println(stateStructv1.GetState().LatestBlockHeader.BodyRoot)
	tmplog.Println(stateStructv2.GetState().LatestBlockHeader.BodyRoot)

	tmplog.Println(stateStructv1.GetState().LatestBlockHeader.StateRoot)
	tmplog.Println(stateStructv2.GetState().LatestBlockHeader.StateRoot)
}

func mkBeaconStateAltair(bytearray []byte) state.BeaconState {
	ethpbState := &ethpb.BeaconStateAltair{}
	err := ethpbState.UnmarshalSSZ(bytearray)
	panicErr(err)

	state, err := statev2.InitializeFromProto(ethpbState)
	panicErr(err)

	return state
}

// Use the blocks to build light-client-updates
func (s *Service) buildLightClientUpdates(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconStateAltair) error {
	//verify_serder_same_root(ctx, state)

	// Header
	header, err := block.Header()
	if err != nil {
		return err
	}

	//currentCom := state.CurrentSyncCommittee()
	nextCom, err := state.NextSyncCommittee()
	if err != nil {
		return err
	}
	//tmplog.Println(nextCom)
	// TODO: ?? how to build committeebranch?

	fCheckpoint := state.FinalizedCheckpoint()

	finalityBlock, err := s.cfg.Database.Block(ctx, bytesutil.ToBytes32(fCheckpoint.Root))
	if err != nil {
		return err
	}
	finalityHeader, err := finalityBlock.Header()
	if err != nil {
		return err
	}
	// TODO: ?? how to build finality header branch?

	//nextSyncCommittee, err := block.Block().Body().SyncAggregate()
	syncAgg, err := block.Block().Body().SyncAggregate()
	if err != nil {
		return err
	}

	update := &ethpb.LightClientUpdate{
		//Header                  *BeaconBlockHeader                                `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
		Header: header.Header,

		//NextSyncCommittee       *SyncCommittee                                    `protobuf:"bytes,2,opt,name=next_sync_committee,json=nextSyncCommittee,proto3" json:"next_sync_committee,omitempty"`
		//NextSyncCommitteeBranch [][]byte                                          `protobuf:"bytes,3,rep,name=next_sync_committee_branch,json=nextSyncCommitteeBranch,proto3" json:"next_sync_committee_branch,omitempty" ssz-size:"5,32"`
		NextSyncCommittee: nextCom,

		//FinalityHeader          *BeaconBlockHeader                                `protobuf:"bytes,4,opt,name=finality_header,json=finalityHeader,proto3" json:"finality_header,omitempty"`
		//FinalityBranch          [][]byte                                          `protobuf:"bytes,5,rep,name=finality_branch,json=finalityBranch,proto3" json:"finality_branch,omitempty" ssz-size:"6,32"`
		FinalityHeader: finalityHeader.Header,

		//SyncCommitteeBits       github_com_prysmaticlabs_go_bitfield.Bitvector512 `protobuf:"bytes,6,opt,name=sync_committee_bits,json=syncCommitteeBits,proto3" json:"sync_committee_bits,omitempty" cast-type:"github.com/prysmaticlabs/go-bitfield.Bitvector512" ssz-size:"64"`
		//SyncCommitteeSignature  []byte                                            `protobuf:"bytes,7,opt,name=sync_committee_signature,json=syncCommitteeSignature,proto3" json:"sync_committee_signature,omitempty" ssz-size:"96"`
		SyncCommitteeBits:      syncAgg.SyncCommitteeBits,
		SyncCommitteeSignature: syncAgg.SyncCommitteeSignature,

		//ForkVersion             []byte                                            `protobuf:"bytes,8,opt,name=fork_version,json=forkVersion,proto3" json:"fork_version,omitempty" ssz-size:"4"`
		ForkVersion: state.Fork().CurrentVersion,
	}

	//// Produce update ssz
	//updateHash, err := update.HashTreeRoot()
	//if err != nil {
	//	panic(err)
	//}
	//filename = "/tmp/ssz/lighclientupdate/" + string(updateHash[:]) + ".ssz"
	//tmplog.Println("write a ssz file:", filename)
	//var updateSsz []byte
	//err = update.UnmarshalSSZ(updateSsz)
	//if err != nil {
	//	panic(err)
	//}
	//err = os.WriteFile(filename, updateSsz, 0666)
	//if err != nil {
	//	panic(err)
	//}

	// build LightClientUpdate
	//update.
	//	tmplog.Println(*update)

	s.Queue.Enqueue(update)
	tmplog.Println(s.Queue.Len(), s.Queue.Cap())

	// Build SkipSyncUpdate

	// Use the block to build a light-client-update Queue

	return nil
}

// Use the blocks to build light-client-updates
func (s *Service) onFinalizedCheckpoint(ctx context.Context, block block.SignedBeaconBlock, state state.BeaconStateAltair) error {
	//blockCopy := block.Copy()
	// build LightClientUpdate
	tmplog.Println(s.Queue.Len(), s.Queue.Cap())

	return nil
}
