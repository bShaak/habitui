all: build

build:
	mkdir -p bin
	go build -o bin/habitui ./cmd/habitui

run:
	go run ./cmd/habitui

test:
	go test ./...

install:
	go install ./cmd/habitui

clean:
	rm -rf bin/

.PHONY: all build run test install clean
