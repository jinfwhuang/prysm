package sync

import (
	"encoding/base64"
	"errors"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/encoding/ssz/ztype"
	tmplog "log"

	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

const GenesisValidatorsRootBase64Str = "SzY9uU4oYSDXbrkFNA/dTlS/6fBr8z/2z1rSf1Eb/pU="

var GenesisValidatorsRoot [32]byte

func init() {
	b, err := base64.StdEncoding.DecodeString(GenesisValidatorsRootBase64Str)
	if err != nil {
		panic(err)
	}
	GenesisValidatorsRoot = bytesutil.ToBytes32(b)
}

// TODO: refactor
func verifyMerkleFinalityHeader(update *ethpb.LightClientUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 20, 1) // next_sync_committee  23
	root := update.Header.StateRoot
	leaf, err := update.FinalityHeader.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin
	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.FinalityBranch
	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

func verifyMerkleNextSyncComm(update *ethpb.LightClientUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 23) // next_sync_committee  23
	root := update.Header.StateRoot
	leaf, err := update.NextSyncCommittee.HashTreeRoot() // ??? Is this hashroot correct? fastszz issue?  TODO: jin

	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.NextSyncCommitteeBranch
	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

func verifyMerkleCurrentSyncComm(update *ethpb.SkipSyncUpdate) bool {
	gIndex := ztype.CalculateGIndex(ztype.BeaconStateAltairType, 22) // current_sync_committee  22
	root := update.Header.StateRoot
	leaf, err := update.CurrentSyncCommittee.HashTreeRoot()

	if err != nil {
		tmplog.Println(err)
		panic(err)
	}
	branch := update.CurrentSyncCommitteeBranch
	return ztype.Verify(bytesutil.ToBytes32(root), gIndex, leaf, branch)
}

func validateMerkleSkipSyncUpdate(update *ethpb.SkipSyncUpdate) error {
	verified := verifyMerkleFinalityHeader(toLightClientUpdate(update))
	verified = verified && verifyMerkleCurrentSyncComm(update)
	verified = verified && verifyMerkleNextSyncComm(toLightClientUpdate(update))
	if !verified {
		return errors.New("update has invalid merkle proof")
	}
	return nil
}

func validateMerkleAll(update *ethpb.LightClientUpdate) error {
	verified := verifyMerkleFinalityHeader(update)
	verified = verified && verifyMerkleNextSyncComm(update)
	if !verified {
		return errors.New("update has invalid merkle proof")
	}
	return nil
}
