.PHONY: proto
proto:
	./scripts/generate-buf.sh

.PHONY: dev-api
dev-api:
	@cd go &&  ENV_FILE=../.env.local ../scripts/envlocal go run cmd/main.go
