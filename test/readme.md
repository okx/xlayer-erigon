# run test
```shell
make build-docker;
cd test; 
make run;

cast send -f 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266  --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --value 0.01ether 0xA6f7A6b2E9B4d41C582D4Aaf907F45321e2Ca847 --legacy --gas-price 1000000000  --rpc-url http://127.0.0.1:8123
```

# todo 
- validium mode must be upgraded.
- agg can't send prove, sync and agg must be upgraded.
