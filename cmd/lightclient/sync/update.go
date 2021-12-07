package sync

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
	"github.com/prysmaticlabs/prysm/crypto/bls/common"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	tmplog "log"
	//v1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	eth2_types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/time/slots"
	"google.golang.org/protobuf/proto"
)

func toLightClientUpdate(update *ethpb.SkipSyncUpdate) *ethpb.LightClientUpdate {
	tmplog.Println(update)
	return &ethpb.LightClientUpdate{
		Header:                  update.Header,
		NextSyncCommittee:       update.NextSyncCommittee,
		NextSyncCommitteeBranch: update.NextSyncCommitteeBranch,
		FinalityHeader:          update.FinalityHeader,
		FinalityBranch:          update.FinalityBranch,
		SyncCommitteeBits:       update.SyncCommitteeBits,
		SyncCommitteeSignature:  update.SyncCommitteeSignature,
		ForkVersion:             update.ForkVersion,
	}
}

// TODO: refactor
func verifyFinalityBranch(update *ethpb.LightClientUpdate) bool {
	emptyZtypeState := ztype.NewEmptyBeaconState()

	root := update.FinalityHeader.StateRoot
	gIndex := emptyZtypeState.GetGIndex(20, 1) // finalized_checkpoint (20), root (1)
	leaf, err := update.Header.HashTreeRoot()  // ??? Is this hashroot correct? fastszz issue?
	if err != nil {
		tmplog.Println(err)
	}
	branch := update.FinalityBranch

	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

// TODO: ??
func verifyNextSyncCommProof(update *ethpb.LightClientUpdate) bool {
	return true
}

// See: https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/sync-protocol.md#validate_light_client_update
func validateLightClientUpdate(
	snapshot *ethpb.LightClientSnapshot,
	update *ethpb.LightClientUpdate,
	genesisValidatorsRoot [32]byte,
) error {
	if update.Header.Slot <= snapshot.Header.Slot {
		return errors.New("wrong")
	}
	snapshotPeriod := slots.ToEpoch(snapshot.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	updatePeriod := slots.ToEpoch(update.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	if updatePeriod != snapshotPeriod || updatePeriod != snapshotPeriod+1 {
		return errors.New("unwanted")
	}

	// Verify finality headers.
	var signedHeader *ethpb.BeaconBlockHeader
	if isEmptyBlockHeader(update.FinalityHeader) {
		signedHeader = update.Header
		// Check if branch is empty.
		for _, elem := range update.FinalityBranch {
			if len(elem) != 0 { // TODO: potential error
				return errors.New("branch not empty")
			}
		}
	} else {
		signedHeader = update.FinalityHeader
		if !verifyFinalityBranch(update) {
			return errors.New("finality branch does not verify")
		}
	}

	// Verify update next sync committee if the update period incremented.
	var syncCommittee *ethpb.SyncCommittee
	if updatePeriod == snapshotPeriod {
		syncCommittee = snapshot.CurrentSyncCommittee
		for _, elem := range update.NextSyncCommitteeBranch {
			if len(elem) != 0 { // TODO: potential error
				return errors.New("branch not empty")
			}
		}
	} else {
		syncCommittee = snapshot.NextSyncCommittee
		verifyNextSyncCommProof(update)
	}

	// Verify sync committee has sufficient participants
	if update.SyncCommitteeBits.Count() < params.BeaconConfig().MinSyncCommitteeParticipants {
		return errors.New("insufficient participants")
	}

	// Verify sync committee aggregate signature
	participantPubkeys := make([][]byte, 0)
	for i, pubKey := range syncCommittee.Pubkeys {
		bit := update.SyncCommitteeBits.BitAt(uint64(i))
		if bit {
			participantPubkeys = append(participantPubkeys, pubKey)
		}
	}
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainSyncCommittee,
		update.ForkVersion,
		genesisValidatorsRoot[:],
	)
	if err != nil {
		return err
	}
	signingRoot, err := signing.ComputeSigningRoot(signedHeader, domain)
	if err != nil {
		return err
	}
	sig, err := blst.SignatureFromBytes(update.SyncCommitteeSignature[:])
	if err != nil {
		return err
	}
	pubKeys := make([]common.PublicKey, 0)
	for _, pubkey := range participantPubkeys {
		pk, err := blst.PublicKeyFromBytes(pubkey)
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pk)
	}
	if !sig.FastAggregateVerify(pubKeys, signingRoot) {
		return errors.New("failed to verify")
	}
	return nil
}

func isEmptyBlockHeader(header *ethpb.BeaconBlockHeader) bool {
	emptyRoot := params.BeaconConfig().ZeroHash
	return proto.Equal(header, &ethpb.BeaconBlockHeader{
		Slot:          0,
		ProposerIndex: 0,
		ParentRoot:    emptyRoot[:],
		StateRoot:     emptyRoot[:],
		BodyRoot:      emptyRoot[:],
	})
}

func applyHeaderUpdate(snapshot *ethpb.LightClientSnapshot, header *ethpb.BeaconBlockHeader) {
	snapshotPeriod := slots.ToEpoch(snapshot.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	updatePeriod := slots.ToEpoch(header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	if updatePeriod == snapshotPeriod+1 {
		snapshot.CurrentSyncCommittee = snapshot.NextSyncCommittee
	} else {
		snapshot.Header = header
	}
}

/**
def process_light_client_update(Store: LightClientStore, update: LightClientUpdate, current_slot: Slot,
                                genesis_validators_root: Root) -> None:
*/
func processLightClientUpdate(
	store *ethpb.LightClientStore,
	update *ethpb.LightClientUpdate,
	currentSlot eth2_types.Slot,
	genesisValidatorsRoot [32]byte,
) error {
	store.Updates = append(store.Updates, update)
	updateTimeout := uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod)
	sumParticipantBits := update.SyncCommitteeBits.Count()
	quorumTwoThird := sumParticipantBits >= update.SyncCommitteeBits.Len()*2.0/3 // 2/3 quorum is reached
	if quorumTwoThird && !isEmptyBlockHeader(update.FinalityHeader) {
		// Only use a finality header
		applyHeaderUpdate(store.Snapshot, update.FinalityHeader) // TODO: There could be a different update algorithm

		// Blindly accepts the latest header
		applyHeaderUpdate(store.Snapshot, update.Header) // TODO: There could be a different update algorithm
		store.Updates = make([]*ethpb.LightClientUpdate, 0)
	} else if currentSlot > store.Snapshot.Header.Slot.Add(updateTimeout) {
		// TODO: use skip-sync
		panic("Not implemented")
		//// Forced best update when the update timeout has elapsed
		////// Use the update that has the highest sum of sync committee bits.
		////updateWithHighestSumBits := Store.Updates[0]
		////highestSumBitsUpdate := updateWithHighestSumBits.SyncCommitteeBits.Count()
		////for _, validUpdate := range Store.Updates {
		////	sumUpdateBits := validUpdate.SyncCommitteeBits.Count()
		////	if sumUpdateBits > highestSumBitsUpdate {
		////		highestSumBitsUpdate = sumUpdateBits
		////		updateWithHighestSumBits = validUpdate
		////	}
		////}
		//bestUpdate := Store.Updates[0] // TODO: hack
		//applyLightClientUpdate(Store.Snapshot, bestUpdate)
		//Store.Updates = make([]*ethpb.LightClientUpdate, 0)
	}
	return nil
}
