package ztype

import (
	"encoding/binary"
	///Users/jin/code/repos/prysm/beacon-chain/state/v2

	//"bytes"
	//"github.com/protolambda/ztyp/codec"
	//"github.com/prysmaticlabs/go-bitfield"
	//"github.com/prysmaticlabs/prysm/beacon-chain/state"
	//"github.com/prysmaticlabs/prysm/config/params"
	"github.com/prysmaticlabs/prysm/io/file"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"strconv"

	//"github.com/prysmaticlabs/prysm/testing/util"
	tmplog "log"
	"testing"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

func TestAA(t *testing.T) {
	tmplog.Println("fff")
}

func Test_HashTreeRoot(t *testing.T) {
	tmplog.Println("fff")

}

func Test_StateConversion(t *testing.T) {
	tmplog.Println("fff")
	//st, _ := util.NewBeaconState()
	//dec := codec.NewDecodingReader(bytes.NewReader(data), uint64(len(data)))
	convertState()
}

func getReadableHash(hash [32]byte) string {
	return strconv.Itoa(int(binary.BigEndian.Uint32(hash[:])))
}

func convertState() *ethpb.BeaconStateAltair {
	//filename := "/Users/jin/code/repos/prysm/tmp/ssz/statev1/64337.ssz"
	//filename := "/Users/jin/code/repos/prysm/tmp/ssz/statev1/48483.ssz"
	filename := "/Users/jin/code/repos/prysm/tmp/ssz/statev1/24691.ssz" // The fastssz hash is: 62113.ssz

	tmplog.Println(filename)

	data, err := file.ReadFileAsBytes(filename)
	if err != nil {
		panic(err)
	}

	// fastssz BeaconState
	fastsszBeaconState := &ethpb.BeaconState{}
	err = fastsszBeaconState.UnmarshalSSZ(data)
	fastsszBeaconStateHash, err := fastsszBeaconState.HashTreeRoot()
	fastsszBeaconStateInt := getReadableHash(fastsszBeaconStateHash)

	// fastssz BeaconStateAltair
	fastsszBeaconStateAltair := &ethpb.BeaconStateAltair{}
	err = fastsszBeaconStateAltair.UnmarshalSSZ(data)
	fastsszBeaconStateAltairHash, err := fastsszBeaconStateAltair.HashTreeRoot()
	fastsszBeaconStateAltairHashInt := getReadableHash(fastsszBeaconStateAltairHash)
	//if err != nil {
	//	panic(err)
	//}

	// prysm BeaconStateAltair

	tmplog.Println(err)
	//if err != nil {
	//	panic(err)
	//}

	tmplog.Println(fastsszBeaconStateAltairHashInt)
	tmplog.Println(fastsszBeaconStateInt)
	//tmplog.Println(stateV1)

	return nil
}
