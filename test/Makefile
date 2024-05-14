.PHONY: run
run:
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose down

.PHONY: sendtx
sendtx:
	cast send --legacy --from 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --rpc-url http://127.0.0.1:8123 0x4A24d2790E265d18FDB68a03703dAde96F486b73 --value 0.00001ether --gas-price=1000000000
