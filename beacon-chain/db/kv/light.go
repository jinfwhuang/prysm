package kv

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	bolt "go.etcd.io/bbolt"
	tmplog "log"
)

func (s *Store) GetSkipSyncUpdate(ctx context.Context, key [32]byte) (*ethpb.SkipSyncUpdate, error) {
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(skipSyncBucket)
		stBytes := bkt.Get(key[:])
		if len(stBytes) == 0 {
			return nil
		}
		// Due to https://github.com/boltdb/bolt/issues/204, we need to
		// allocate a byte slice separately in the transaction or there
		// is the possibility of a panic when accessing that particular
		// area of memory.
		value = make([]byte, len(stBytes))
		copy(value, stBytes)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		err := fmt.Errorf("cannot find skip sync error, 0xkey=%s, base64key=%s",
			hex.EncodeToString(key[:]),
			base64.StdEncoding.EncodeToString(key[:]),
		)
		return nil, err
	}

	update := &ethpb.SkipSyncUpdate{}
	update.UnmarshalSSZ(value)
	return update, nil
}

func (s *Store) SaveSkipSyncUpdate(ctx context.Context, update *ethpb.SkipSyncUpdate) error {
	key, _ := update.CurrentSyncCommittee.HashTreeRoot()
	value, _ := update.MarshalSSZ()
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(skipSyncBucket)
		if err := bucket.Put(key[:], value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	updateRoot, _ := update.HashTreeRoot()
	tmplog.Println("key base64", base64.StdEncoding.EncodeToString(key[:]), "skipsyncupdate hash root", hex.EncodeToString(updateRoot[:]))
	return nil
}
