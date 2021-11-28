package ztype

import (
	"bytes"
	"github.com/protolambda/ztyp/codec"
	//"bytes"
	"sort"

	//"github.com/pkg/errors"
	//"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
	//"github.com/prysmaticlabs/prysm/beacon-chain/state"
	//"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	v2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
)

var (
	BLSPubkeyType = view.BasicVectorType(view.ByteType, 48)
	ValidatorType = view.ContainerType("Validator", []view.FieldDef{
		{"pubkey", BLSPubkeyType},
		{"withdrawal_credentials", view.RootType},
		{"effective_balance", view.Uint64Type},
		{"slashed", view.BoolType},
		{"activation_eligibility_epoch", view.Uint64Type},
		{"activation_epoch", view.Uint64Type},
		{"exit_epoch", view.Uint64Type},
		{"withdrawable_epoch", view.Uint64Type},
	})
	ForkType = view.ContainerType("Fork", []view.FieldDef{
		{"previous_version", view.Bytes4Type},
		{"current_version", view.Bytes4Type},
		{"epoch", view.Uint64Type},
	})
	BeaconBlockHeaderType = view.ContainerType("BeaconBlockHeader", []view.FieldDef{
		{"slot", view.Uint64Type},
		{"proposer_index", view.Uint64Type},
		{"parent_root", view.RootType},
		{"state_root", view.RootType},
		{"body_root", view.RootType},
	})
	BlockRootsType      = view.VectorType(view.RootType, 8192)
	StateRootsType      = view.VectorType(view.RootType, 8192)
	HistoricalRootsType = view.ListType(view.RootType, 16777216)
	Eth1DataType        = view.ContainerType("Eth1Data", []view.FieldDef{
		{"deposit_root", view.RootType},
		{"deposit_count", view.Uint64Type},
		{"block_hash", view.RootType},
	})
	Eth1DataVotesType     = view.ComplexListType(Eth1DataType, 2048)
	ValidatorsType        = view.ComplexListType(ValidatorType, 1099511627776)
	BalancesType          = view.BasicListType(view.Uint64Type, 1099511627776)
	RandaoMixesType       = view.VectorType(view.RootType, 65536)
	SlashingsType         = view.BasicVectorType(view.Uint64Type, 8192)
	ParticipationType     = view.BasicListType(view.ByteType, 1099511627776)
	JustificationBitsType = view.BitVectorType(4)
	CheckpointType        = view.ContainerType("Checkpoint", []view.FieldDef{
		{"epoch", view.Uint64Type},
		{"root", view.RootType},
	})
	InactivityScoresType  = view.BasicListType(view.Uint64Type, 1099511627776)
	SyncCommitteeKeysType = view.VectorType(BLSPubkeyType, 512)
	SyncCommitteeType     = view.ContainerType("SyncCommittee", []view.FieldDef{
		{"pubkeys", SyncCommitteeKeysType},
		{"aggregate_pubkey", BLSPubkeyType},
	})
	BeaconStateAltairType = view.ContainerType("BeaconStateAltair", []view.FieldDef{
		{"genesis_time", view.Uint64Type}, // 0
		{"genesis_validators_root", view.RootType},
		{"slot", view.Uint64Type},
		{"fork", ForkType},
		{"latest_block_header", BeaconBlockHeaderType},
		{"block_roots", BlockRootsType}, // 5
		{"state_roots", StateRootsType},
		{"historical_roots", HistoricalRootsType},
		{"eth1_data", Eth1DataType},
		{"eth1_data_votes", Eth1DataVotesType},
		{"eth1_deposit_index", view.Uint64Type}, // 10
		{"validators", ValidatorsType},
		{"balances", BalancesType},
		{"randao_mixes", RandaoMixesType},
		{"slashings", SlashingsType},
		{"previous_epoch_participation", ParticipationType}, // 15
		{"current_epoch_participation", ParticipationType},
		{"justification_bits", JustificationBitsType},
		{"previous_justified_checkpoint", CheckpointType},
		{"current_justified_checkpoint", CheckpointType},
		{"finalized_checkpoint", CheckpointType}, // 20
		{"inactivity_scores", InactivityScoresType},
		{"current_sync_committee", SyncCommitteeType},
		{"next_sync_committee", SyncCommitteeType}, // 23
	})
)

type ZType interface {
	HashTreeRoot() tree.Root
	SszSerialize() []byte
	Proof(index tree.Gindex64) (leave tree.Root, branch [][]byte)
	Verify(leave tree.Root, branch [][]byte, index tree.Gindex64, root tree.Root) bool
}

type TreeBaseBeaconStateAltair struct {
	//def  view.ContainerTypeDef
	view *view.ContainerView
}

func NewTreeBaseBeaconStateAltair(sszBytes []byte) *TreeBaseBeaconStateAltair {
	dec := codec.NewDecodingReader(bytes.NewReader(sszBytes), uint64(len(sszBytes)))
	stateView, err := BeaconStateAltairType.Deserialize(dec)

	//view.View()
	if err != nil {
		panic(err)
	}
	//cc := a.(*view.ContainerView)
	//
	return &TreeBaseBeaconStateAltair{
		view: stateView.(*view.ContainerView),
	}
}

func NewTreeBackedState(state v2.BeaconState) *TreeBaseBeaconStateAltair {
	ser, err := state.MarshalSSZ()
	if err != nil {
		panic(err)
	}
	reader := codec.NewDecodingReader(bytes.NewReader(ser), uint64(len(ser)))
	stateView, err := BeaconStateAltairType.Deserialize(reader)

	return &TreeBaseBeaconStateAltair{
		view: stateView.(*view.ContainerView),
	}
}

// GetGIndex Get the generalized index of the field
func (obj *TreeBaseBeaconStateAltair) GetGIndex(fieldIndex uint64) tree.Gindex64 {
	return 0
}

func (obj *TreeBaseBeaconStateAltair) HashTreeRoot() tree.Root {
	return [32]byte{}
}

func (obj *TreeBaseBeaconStateAltair) SszSerialize() []byte {
	return nil
}

/*
def verify_merkle_proof(leaf: Bytes32, proof: Sequence[Bytes32], index: GeneralizedIndex, root: Root) -> bool:
    return calculate_merkle_root(leaf, proof, index) == root
*/
func (obj *TreeBaseBeaconStateAltair) Verify(leave tree.Root, branch [][]byte, index tree.Gindex64, root tree.Root) bool {
	panic("not implemented")
}

func (obj *TreeBaseBeaconStateAltair) Proof(index tree.Gindex64) (leave tree.Root, branch [][]byte) {
	// TODO: jin fix the type assumption
	//depth := tree.CoverDepth(obj.view.FieldCount())
	//index
	//generalizedIdx, _ = tree.ToGindex64(fieldIndex, depth)
	//if err != nil {
	//	return
	//}
	leaves := make(map[tree.Gindex64]struct{})
	leaves[index] = struct{}{}
	leavesSorted := make([]tree.Gindex64, 0, len(leaves))
	for g := range leaves {
		leavesSorted = append(leavesSorted, g)
	}
	sort.Slice(leavesSorted, func(i, j int) bool {
		return leavesSorted[i] < leavesSorted[j]
	})

	// Mark every gindex that is between the root and the leaves.
	interest := make(map[tree.Gindex64]struct{})
	for _, g := range leavesSorted {
		iter, _ := g.BitIter()
		n := tree.Gindex64(1)
		for {
			right, ok := iter.Next()
			if !ok {
				break
			}
			n *= 2
			if right {
				n += 1
			}
			interest[n] = struct{}{}
		}
	}
	witness := make(map[tree.Gindex64]struct{})
	// For every gindex that is covered, check if the sibling is covered, and if not, it's a witness
	for g := range interest {
		if _, ok := interest[g^1]; !ok {
			witness[g^1] = struct{}{}
		}
	}
	witnessSorted := make([]tree.Gindex64, 0, len(witness))
	for g := range witness {
		witnessSorted = append(witnessSorted, g)
	}
	sort.Slice(witnessSorted, func(i, j int) bool {
		return witnessSorted[i] < witnessSorted[j]
	})

	node := obj.view.BackingNode
	hFn := tree.GetHashFn()
	branch = make([][]byte, 0, len(witnessSorted))
	for i := len(witnessSorted) - 1; i >= 0; i-- {
		g := witnessSorted[i]
		n, err := node.Getter(g)
		if err != nil {
			panic(err)
		}
		root := n.MerkleRoot(hFn)
		branch = append(branch, root[:])
	}
	return [32]byte{}, branch
	//return nil, 0
}
