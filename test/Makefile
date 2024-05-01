.PHONY: run
run:
	docker-compose up -d xlayer-aggregator-db
	docker-compose up -d xlayer-state-db
	docker-compose up -d xlayer-mock-l1-network
	docker-compose up -d xlayer-stateless-executor
	docker-compose up -d xlayer-prover
	sleep 3
	docker-compose up -d xlayer-erigon
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
	docker-compose stop xlayer-erigon
	docker-compose stop xlayer-seqsender
	docker-compose stop xlayer-aggregator

.PHONY: remove
remove:
	docker-compose remove xlayer-aggregator-db
	docker-compose remove xlayer-state-db
	docker-compose remove xlayer-mock-l1-network
	docker-compose remove xlayer-stateless-executor
	docker-compose remove xlayer-prover
	docker-compose remove xlayer-erigon
	docker-compose remove xlayer-seqsender
	docker-compose remove xlayer-aggregator



