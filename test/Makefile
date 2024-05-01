.PHONY: run
run:
	docker-compose up -d zkevm-aggregator-db
	docker-compose up -d zkevm-state-db
	docker-compose up -d zkevm-mock-l1-network
	docker-compose up -d zkevm-stateless-executor
	docker-compose up -d zkevm-prover
	sleep 3
	docker-compose up -d zkevm-erigon
	sleep 3
	docker-compose up -d zkevm-seqsender
	docker-compose up -d zkevm-aggregator

.PHONY: stop
stop:
	docker-compose stop zkevm-aggregator-db
	docker-compose stop zkevm-state-db
	docker-compose stop zkevm-mock-l1-network
	docker-compose stop zkevm-stateless-executor
	docker-compose stop zkevm-prover
	docker-compose stop zkevm-erigon
	docker-compose stop zkevm-seqsender
	docker-compose stop zkevm-aggregator

.PHONY: remove
remove:
	docker-compose remove zkevm-aggregator-db
	docker-compose remove zkevm-state-db
	docker-compose remove zkevm-mock-l1-network
	docker-compose remove zkevm-stateless-executor
	docker-compose remove zkevm-prover
	docker-compose remove zkevm-erigon
	docker-compose remove zkevm-seqsender
	docker-compose remove zkevm-aggregator



