DELVE_VERSION := 1.23.1

default: test lint run

.PHONY: test
test: ARGS=
test:
	go test -v $(ARGS) .

.PHONY: lint
lint:
	docker compose run --rm golangci

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

.PHONY: install-delve
install-delve:
	go install "github.com/go-delve/delve/cmd/dlv@v$(DELVE_VERSION)"

.PHONY: debug-start
debug-start: install-delve
	dlv debug --headless --listen 'localhost:4040' .

.PHONY: debug-connect
debug-connect: install-delve
	dlv connect :4040

.PHONY: debug-test
debug-test: ARGS=
debug-test: install-delve
	dlv test $(ARGS) .
