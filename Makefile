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
run: ARGS=
run:
	go run . $(ARGS)

.PHONY: ./sazed
./sazed:
	go build .

build: ./sazed

.PHONY: install
install: DIR=~/.local/bin
install: build
	mkdir -p $(DIR)
	install -m 744 sazed $(DIR)
