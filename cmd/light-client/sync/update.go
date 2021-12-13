package sync

import (
	//v1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	eth2_types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/config/params"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/time/slots"
)

func toLightClientUpdate(update *ethpb.SkipSyncUpdate) *ethpb.LightClientUpdate {
	return &ethpb.LightClientUpdate{
		AttestedHeader:          update.AttestedHeader,
		NextSyncCommittee:       update.NextSyncCommittee,
		NextSyncCommitteeBranch: update.NextSyncCommitteeBranch,
		FinalityHeader:          update.FinalityHeader,
		FinalityBranch:          update.FinalityBranch,
		SyncCommitteeBits:       update.SyncCommitteeBits,
		SyncCommitteeSignature:  update.SyncCommitteeSignature,
		ForkVersion:             update.ForkVersion,
	}
}

/*
def get_safety_threshold(store: LightClientStore) -> uint64:
    return max(
        store.previous_max_active_participants,
        store.current_max_active_participants
    ) // 2
*/
func getSafetyThreshold(store *ethpb.LightClientStore) uint64 {
	max := store.PreviousMaxActiveParticipants
	if store.CurrentMaxActiveParticipants > max {
		max = store.CurrentMaxActiveParticipants
	}
	return max / 2
}

//// TODO: refactor
//func verifyFinalityBranch(update *ethpb.LightClientUpdate) bool {
//	emptyZtypeState := ztype.NewEmptyBeaconState()
//
//	root := update.FinalityHeader.StateRoot
//	gIndex := emptyZtypeState.GetGIndex(20, 1) // finalized_checkpoint (20), root (1)
//	leaf, err := update.Header.HashTreeRoot()  // ??? Is this hashroot correct? fastszz issue?
//	if err != nil {
//		tmplog.Println(err)
//	}
//	branch := update.FinalityBranch
//
//	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
//}

// TODO: ??
func verifyNextSyncCommProof(update *ethpb.LightClientUpdate) bool {
	return true
}

//// See: https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/sync-protocol.md#validate_light_client_update
//func validateLightClientUpdate(
//	snapshot *ethpb.LightClientSnapshot,
//	update *ethpb.LightClientUpdate,
//	genesisValidatorsRoot [32]byte,
//) error {
//	if update.Header.Slot <= snapshot.Header.Slot {
//		return errors.New("wrong")
//	}
//	snapshotPeriod := slots.ToEpoch(snapshot.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
//	updatePeriod := slots.ToEpoch(update.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
//	if updatePeriod != snapshotPeriod || updatePeriod != snapshotPeriod+1 {
//		return errors.New("unwanted")
//	}
//
//	// Verify finality headers.
//	var signedHeader *ethpb.BeaconBlockHeader
//	if isEmptyBlockHeader(update.FinalityHeader) {
//		signedHeader = update.Header
//		// Check if branch is empty.
//		for _, elem := range update.FinalityBranch {
//			if len(elem) != 0 { // TODO: potential error
//				return errors.New("branch not empty")
//			}
//		}
//	} else {
//		signedHeader = update.FinalityHeader
//		if !verifyFinalityBranch(update) {
//			return errors.New("finality branch does not verify")
//		}
//	}
//
//	// Verify update next sync committee if the update period incremented.
//	var syncCommittee *ethpb.SyncCommittee
//	if updatePeriod == snapshotPeriod {
//		syncCommittee = snapshot.CurrentSyncCommittee
//		for _, elem := range update.NextSyncCommitteeBranch {
//			if len(elem) != 0 { // TODO: potential error
//				return errors.New("branch not empty")
//			}
//		}
//	} else {
//		syncCommittee = snapshot.NextSyncCommittee
//		verifyNextSyncCommProof(update)
//	}
//
//	// Verify sync committee has sufficient participants
//	if update.SyncCommitteeBits.Count() < params.BeaconConfig().MinSyncCommitteeParticipants {
//		return errors.New("insufficient participants")
//	}
//
//	// Verify sync committee aggregate signature
//	participantPubkeys := make([][]byte, 0)
//	for i, pubKey := range syncCommittee.Pubkeys {
//		bit := update.SyncCommitteeBits.BitAt(uint64(i))
//		if bit {
//			participantPubkeys = append(participantPubkeys, pubKey)
//		}
//	}
//	domain, err := signing.ComputeDomain(
//		params.BeaconConfig().DomainSyncCommittee,
//		update.ForkVersion,
//		genesisValidatorsRoot[:],
//	)
//	if err != nil {
//		return err
//	}
//	signingRoot, err := signing.ComputeSigningRoot(signedHeader, domain)
//	if err != nil {
//		return err
//	}
//	sig, err := blst.SignatureFromBytes(update.SyncCommitteeSignature[:])
//	if err != nil {
//		return err
//	}
//	pubKeys := make([]common.PublicKey, 0)
//	for _, pubkey := range participantPubkeys {
//		pk, err := blst.PublicKeyFromBytes(pubkey)
//		if err != nil {
//			return err
//		}
//		pubKeys = append(pubKeys, pk)
//	}
//	if !sig.FastAggregateVerify(pubKeys, signingRoot) {
//		return errors.New("failed to verify")
//	}
//	return nil
//}

//func isEmptyBlockHeader(header *ethpb.BeaconBlockHeader) bool {
//	emptyRoot := params.BeaconConfig().ZeroHash
//	return proto.Equal(header, &ethpb.BeaconBlockHeader{
//		Slot:          0,
//		ProposerIndex: 0,
//		ParentRoot:    emptyRoot[:],
//		StateRoot:     emptyRoot[:],
//		BodyRoot:      emptyRoot[:],
//	})
//}

//func processLightClientUpdate(
//	store *ethpb.LightClientStore,
//	update *ethpb.LightClientUpdate,
//	genesisValidatorsRoot [32]byte,
//) error {
//	// Validations
//	validateMerkleLightClientUpdate(update)
//
//	if IsEmptyHeader(update.FinalityHeader) { // non-finalized
//		/**
//		1.
//		  # Verify update slot is larger than slot of current best finalized header
//		  active_header = get_active_header(update)
//		  assert current_slot >= active_header.slot > store.finalized_header.slot
//		*/
//		root := update.AttestedHeader.StateRoot
//		currentSlot
//		if
//
//	} else { // finalized
//
//	}
//
//	return nil
//}

//func applyHeaderUpdate(snapshot *ethpb.LightClientSnapshot, header *ethpb.BeaconBlockHeader) {
//	snapshotPeriod := slots.ToEpoch(snapshot.Header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
//	updatePeriod := slots.ToEpoch(header.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
//	if updatePeriod == snapshotPeriod+1 {
//		snapshot.CurrentSyncCommittee = snapshot.NextSyncCommittee
//	} else {
//		snapshot.Header = header
//	}
//}

func processLightClientUpdate(
	store *ethpb.LightClientStore,
	update *ethpb.LightClientUpdate,
	currentSlot eth2_types.Slot,
	genesisValidatorsRoot [32]byte,
) error {
	validateLightClientUpdate(store, update, currentSlot, genesisValidatorsRoot)

	/*
	   # Update the best update in case we have to force-update to it if the timeout elapses
	   if (
	       store.best_valid_update is None
	       or sum(update.sync_committee_bits) > sum(store.best_valid_update.sync_committee_bits)
	   ):
	       store.best_valid_update = update

	*/
	// 1. Update the best update in case we have to force-update to it if the timeout elapses
	if store.BestValidUpdate == nil ||
		update.SyncCommitteeBits.Count() > store.BestValidUpdate.SyncCommitteeBits.Count() {
		store.BestValidUpdate = update
	}
	/*
	  # Track the maximum number of active participants in the committee signatures
	  store.current_max_active_participants = max(
	      store.current_max_active_participants,
	      sum(update.sync_committee_bits),
	*/
	// 2. Track the maximum number of active participants in the committee signatures
	if update.SyncCommitteeBits.Count() > store.CurrentMaxActiveParticipants {
		store.CurrentMaxActiveParticipants = update.SyncCommitteeBits.Count()
	}

	/*
	   # Update the optimistic header
	   if (
	       sum(update.sync_committee_bits) > get_safety_threshold(store) and
	       update.attested_header.slot > store.optimistic_header.slot
	   ):
	       store.optimistic_header = update.attested_header
	*/

	// 3. Update the optimistic header
	if update.SyncCommitteeBits.Count() > getSafetyThreshold(store) &&
		update.AttestedHeader.Slot > store.OptimisticHeader.Slot {
		store.OptimisticHeader = update.AttestedHeader
	}

	/*
	   # Update finalized header
	   if (
	       sum(update.sync_committee_bits) * 3 >= len(update.sync_committee_bits) * 2
	       and update.finalized_header != BeaconBlockHeader()
	   ):
	       # Normal update through 2/3 threshold
	       apply_light_client_update(store, update)
	       store.best_valid_update = None
	   elif (
	       current_slot > store.finalized_header.slot + UPDATE_TIMEOUT
	       and store.best_valid_update is not None
	   ):
	       # Forced best update when the update timeout has elapsed
	       apply_light_client_update(store, store.best_valid_update)
	       store.best_valid_update = None
	*/
	// 4. Update finalized header
	if !IsEmptyHeader(update.FinalityHeader) &&
		update.SyncCommitteeBits.Count() >= update.SyncCommitteeBits.Len()*3.0/2.0 {
		// Normal update through 2/3 threshold
		applyLightClientUpdate(store, update)
	} else if currentSlot > store.FinalizedHeader.Slot+UpdateTimeout &&
		store.BestValidUpdate != nil {
		applyLightClientUpdate(store, store.BestValidUpdate)
		store.BestValidUpdate = nil
	}

	return nil
}

/**
def apply_light_client_update(store: LightClientStore, update: LightClientUpdate) -> None:
    active_header = get_active_header(update)
    finalized_period = compute_epoch_at_slot(store.finalized_header.slot) // EPOCHS_PER_SYNC_COMMITTEE_PERIOD
    update_period = compute_epoch_at_slot(active_header.slot) // EPOCHS_PER_SYNC_COMMITTEE_PERIOD
    if update_period == finalized_period + 1:
        store.current_sync_committee = store.next_sync_committee
        store.next_sync_committee = update.next_sync_committee
    store.finalized_header = active_header

*/
func applyLightClientUpdate(store *ethpb.LightClientStore, update *ethpb.LightClientUpdate) {
	activeHeader := getActiveHeader(update)
	finalizedPeriod := slots.ToEpoch(store.FinalizedHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	updatePeriod := slots.ToEpoch(activeHeader.Slot) / params.BeaconConfig().EpochsPerSyncCommitteePeriod
	if updatePeriod == finalizedPeriod+1 {
		store.CurrentSyncCommittee = store.NextSyncCommittee
	}
	store.FinalizedHeader = activeHeader
}
