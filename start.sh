#export GOROOT="/usr/local/go1.20"
#export PATH="/usr/local/go1.20/bin/:/Users/oker/go/bin/:/Users/oker/bin:/Users/oker/.cargo/bin:$PATH"
go version

make cdk-erigon
rm /Users/oker/go/bin/cdk-erigon
cp build/bin/cdk-erigon /Users/oker/go/bin

# cardona step 1
#/Users/oker/go/bin/cdk-erigon --zkevm.sync-limit=896191 --config="hermezconfig-cardona.yaml"
/Users/oker/go/bin/cdk-erigon --config="xlayerconfig-testnet.yaml"

# cardona step 2
#CDK_ERIGON_SEQUENCER=1 /Users/oker/go/bin/cdk-erigon --zkevm.l1-sync-start-block=4789190 --config="hermezconfig-cardona.yaml"