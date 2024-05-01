.PHONY: run
run:
	docker-compose up -d xlayer-aggregator-db
	docker-compose up -d xlayer-state-db
	docker-compose up -d xlayer-mock-l1-network
	docker-compose up -d xlayer-stateless-executor
	docker-compose up -d xlayer-prover
	sleep 3
	docker-compose up -d xlayer-erigon-seq
	sleep 3
	docker-compose up -d xlayer-approve
	docker-compose up -d xlayer-seqsender
	docker-compose up -d xlayer-aggregator

.PHONY: stop
stop:
	docker-compose stop xlayer-aggregator-db
	docker-compose stop xlayer-state-db
	docker-compose stop xlayer-mock-l1-network
	docker-compose stop xlayer-stateless-executor
	docker-compose stop xlayer-prover
	docker-compose stop xlayer-erigon-seq
	docker-compose stop xlayer-seqsender
	docker-compose stop xlayer-aggregator

.PHONY: down
down:
	docker-compose down

.PHONY: sendtx
sendtx:
	cast send --legacy --from 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --rpc-url http://127.0.0.1:8123 0x4A24d2790E265d18FDB68a03703dAde96F486b73 --value 0.00001ether





