default: test fmt lint

.PHONY: test
test:
	go test -v .

.PHONY: fmt
fmt:
	go fmt .

.PHONY: lint
lint:
	docker compose run golangci

.PHONY: run
run:
	go run .
