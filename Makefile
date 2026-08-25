all: build

build:
	mkdir -p bin
	go build -o bin/habitui ./cmd/habitui

run:
	go run ./cmd/habitui

serve:
	go run ./cmd/habitui serve

test:
	go test ./...

vet:
	go vet ./...

install:
	go install ./cmd/habitui

clean:
	rm -rf bin/

.PHONY: all build run serve test vet install clean
