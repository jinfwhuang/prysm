

## Dev

```bash
# Start a server that supports the APIs needed by a light client
go run ./cmd/beacon-chain --datadir=../prysm-data/mainnet --http-web3provider=https://mainnet.infura.io/v3/ecfa17010caa47a0afd0a543c43bbcc3

# Check the necessary APIs are available 
grpcurl -plaintext localhost:4000 list ethereum.eth.v1alpha1.LightClient
ethereum.eth.v1alpha1.LightClient.GetSkipSyncUpdate
ethereum.eth.v1alpha1.LightClient.GetUpdates

# Start light client
go run -v ./cmd/lightclient/
```

## Notes
- Move the core light-client code outside of `cmd/.` 
- 