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

.PHONY: ./sazed
./sazed:
	go build .

build: ./sazed

.PHONY: install
install: DIR=~/.local/bin
install: build
	mkdir -p $(DIR)
	install -m 744 sazed $(DIR)
