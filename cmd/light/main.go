package main

import (
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/cmd"
	"github.com/prysmaticlabs/prysm/cmd/beacon-chain/flags"

	//"github.com/prysmaticlabs/prysm/beacon-chain/light"
	lightnode "github.com/prysmaticlabs/prysm/beacon-chain/node/light"
	v1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	v2 "github.com/prysmaticlabs/prysm/proto/eth/v2"
	v1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	tmplog "log"
	"os"
)

func init() {
	tmplog.SetFlags(tmplog.Llongfile)
}

// Precomputed values for generalized indices.
const (
	FinalizedRootIndex              = 105
	FinalizedRootIndexFloorLog2     = 6
	NextSyncCommitteeIndex          = 55
	NextSyncCommitteeIndexFloorLog2 = 5
)

var log = logrus.WithField("prefix", "light")

type LightClientSnapshot struct {
	Header               *v1.BeaconBlockHeader
	CurrentSyncCommittee *v2.SyncCommittee
	NextSyncCommittee    *v2.SyncCommittee
}

type LightClientUpdate struct {
	Header                  *v1.BeaconBlockHeader
	NextSyncCommittee       *v2.SyncCommittee
	NextSyncCommitteeBranch [NextSyncCommitteeIndexFloorLog2][32]byte
	FinalityHeader          *v1.BeaconBlockHeader
	FinalityBranch          [FinalizedRootIndexFloorLog2][32]byte
	SyncCommitteeBits       bitfield.Bitvector512
	SyncCommitteeSignature  [96]byte
	ForkVersion             *v1alpha1.Version
}

type Store struct {
	Snapshot     *LightClientSnapshot
	ValidUpdates []*LightClientUpdate
}

var (
	ServerEndpoint = &cli.StringFlag{
		Name:  "light-client-server-endpoint",
		Usage: "light-client-server-endpoint",
		Value: "localhost:4001", // 512 Mb as a default value.
	}
	ServerEndpointPort = &cli.IntFlag{
		Name:  "light-client-server-endpoint-port",
		Usage: "light-client-server-endpoint port",
		Value: 536870912, // 512 Mb as a default value.
	}

	//GrpcMaxCallRecvMsgSizeFlag = &cli.IntFlag{
	//	Name:  "grpc-max-msg-size",
	//	Usage: "Integer to define max recieve message call size (default: 4194304 (for 4MB))",
	//	Value: 1 << 22,
	//}

)

var appFlags = []cli.Flag{
	ServerEndpoint,
	ServerEndpointPort,
	flags.RPCHostLight,
	flags.RPCPortLight,
	cmd.GrpcMaxCallRecvMsgSizeFlag,
}

func main() {
	app := cli.App{}
	app.Name = "beacon-chain-light-client"
	app.Usage = "Beacon Chain Light Client"
	app.Action = start

	app.Flags = appFlags

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
	}

}

func start(ctx *cli.Context) error {
	// Fix data dir for Windows users.
	//outdatedDataDir := filepath.Join(file.HomeDir(), "AppData", "Roaming", "Eth2")
	serverEndpoint := ctx.String(ServerEndpoint.Name)

	tmplog.Println(serverEndpoint)

	tmplog.Println("staring light client")
	//count := 0
	//for {
	//	tmplog.Println("counting", count)
	//	count += 1
	//	time.Sleep(time.Second * 10)
	//}
	lightNode, err := lightnode.New(ctx)
	if err != nil {
		panic(err)
	}
	lightNode.Start()
	return nil
}
