.PHONY: proto
proto:
	./scripts/generate-buf.sh

.PHONY: dev-api
dev-api:
	@cd go &&  ENV_FILE=../.env.local ../scripts/envlocal go run cmd/main.go

.PHONY: go-test
go-test:
	@cd go && go test ./...

.PHONY:
docker-up:
	docker compose up -d kafka-topics
	docker compose up -d flyway-pg
