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

.PHONY: remove
remove:
	docker stop $(docker ps -aq); docker rm $(docker ps -aq);docker ps -a




