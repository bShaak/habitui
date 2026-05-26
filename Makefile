all: build

build:
	go build -o bin/habitui ./cmd/habitui

run:
	go run ./cmd/habitui

install:
	go install ./cmd/habitui

clean:
	rm -rf bin/
