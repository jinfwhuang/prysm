package kv

import (
	"context"
	"encoding/hex"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	bolt "go.etcd.io/bbolt"
	tmplog "log"
)

//GetSkipSyncUpdate(ctx context.Context, key [32]byte) (error, *ethpb.SkipSyncUpdate)
//SaveSkipSyncUpdate(ctx context.Context, update *ethpb.SkipSyncUpdate) error

// Block retrieval by root.
func (s *Store) GetSkipSyncUpdate(ctx context.Context, key [32]byte) (*ethpb.SkipSyncUpdate, error) {
	tmplog.Println("asking db for look up a skip sync update")
	tmplog.Println("key  :", hex.EncodeToString(key[:]))
	return nil, nil
}

func (s *Store) SaveSkipSyncUpdate(ctx context.Context, update *ethpb.SkipSyncUpdate) error {
	tmplog.Println("asking db to save a skip sync update")

	//bb := &ethpb.SkipSyncUpdate {}
	//bb.

	key, _ := update.CurrentSyncCommittee.HashTreeRoot()
	value, _ := update. // Use ssz serialization
				tmplog.Println("key  :", hex.EncodeToString(key[:]))

	s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(skipSyncBucket)
		if err := bucket.Put(key); err != nil {
			panic(err)
		}
		//for i, rt := range blockRoots {
		//	indicesByBucket := createStateIndicesFromStateSlot(ctx, states[i].Slot())
		//	if err := updateValueForIndices(ctx, indicesByBucket, rt[:], tx); err != nil {
		//		return errors.Wrap(err, "could not update DB indices")
		//	}
		//	if err := bucket.Put(rt[:], multipleEncs[i]); err != nil {
		//		return err
		//	}
		//}
		return nil
	})

	return nil
}
