.PHONY: build run clean test

build:
	go build -o bin/duet .

run: build
	./bin/duet

clean:
	rm -rf bin/

test:
	go test -v ./...

dev:
	go run main.go -worker http://localhost:8787

dev-remote:
	go run main.go -worker https://duet-cf-worker.incident-agent.workers.dev

install:
	go install .
