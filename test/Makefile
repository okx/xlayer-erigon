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
	docker-compose down xlayer-aggregator-db
	docker-compose down xlayer-state-db
	docker-compose down xlayer-mock-l1-network
	docker-compose down xlayer-stateless-executor
	docker-compose down xlayer-prover
	docker-compose down xlayer-erigon-seq
	docker-compose down xlayer-seqsender
	docker-compose down xlayer-aggregator




