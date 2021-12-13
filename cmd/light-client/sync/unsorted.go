package sync

import (
	"github.com/golang/protobuf/proto"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
)

// TODO: use a better filename

func IsEmptyHeader(header *ethpb.BeaconBlockHeader) bool {
	emptyHeader := &ethpb.BeaconBlockHeader{}
	return proto.Equal(header, emptyHeader)
}
