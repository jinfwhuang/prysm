

```bash
grpcurl -plaintext localhost:4000 ethereum.eth.v1alpha1.BeaconChain.GetChainHead

curl 127.0.0.1:3500/eth/v1alpha1/beacon/chainhead | jq

curl 127.0.0.1:3500/eth/v1/beacon/headers | jq

[jin] ~/code/repos/prysm $ grpcurl -plaintext localhost:4000 describe ethereum.eth.service.BeaconChain.ListBlockHeaders
ethereum.eth.service.BeaconChain.ListBlockHeaders is a method:
rpc ListBlockHeaders ( .ethereum.eth.v1.BlockHeadersRequest ) returns ( .ethereum.eth.v1.BlockHeadersResponse ) {
  option (.google.api.http) = { get:"/internal/eth/v1/beacon/headers"  };
}

grpcurl -plaintext localhost:4000 describe ethereum.eth.service.BeaconChain.ListSyncCommittees

rpc ListSyncCommittees ( .ethereum.eth.v2.StateSyncCommitteesRequest ) returns ( .ethereum.eth.v2.StateSyncCommitteesResponse ) {
  option (.google.api.http) = { get:"/internal/eth/v1/beacon/states/{state_id}/sync_committees"  };
}

	StateId []byte
	Epoch   *github_com_prysmaticlabs_eth2_types.Epoch



```

# Get a recent current sys committee hash
```bash
# get a state root
grpcurl -plaintext localhost:4000 ethereum.eth.service.BeaconChain.ListBlockHeaders

# get a sync comm
grpcurl -plaintext \
-d '
{
 "state_id": "ZVYYxMB7/RBft7LdSc3DHDfz4y4MrwFyfVALoJQ+4BA="
}
' \
localhost:4000 ethereum.eth.service.BeaconChain.ListSyncCommittees 

ZVYYxMB7/RBft7LdSc3DHDfz4y4MrwFyfVALoJQ+4BA=


```

https://ethereum.github.io/beacon-APIs/
