.PHONY: build build-linux build-mac build-windows clean

BINARY_NAME=nrc2strava
MAIN_PATH=./cmd/main.go

build:
	go mod tidy
	make build-all

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-linux"

build-mac-intel:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac-intel $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-mac-intel"

build-mac-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-mac-arm $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-mac-arm"

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows.exe $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-windows.exe"

build-all: build-linux build-mac-intel build-mac-arm build-windows

clean:
	rm -rf bin/ go.sum

tidy:
	go mod tidy

.DEFAULT_GOAL := build
