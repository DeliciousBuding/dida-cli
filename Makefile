.PHONY: test build install-local

test:
	go test ./...

build:
	go build -o bin/dida ./cmd/dida

install-local: build
	mkdir -p "$(HOME)/.local/bin"
	cp bin/dida "$(HOME)/.local/bin/dida"

