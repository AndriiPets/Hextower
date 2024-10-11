build:
	@CGO_ENABLED=0 go build -o bin/bit.exe

build_small:
	@CGO_ENABLED=0 go build -ldflags "-s -w -buildid=" -trimpath -o bin/bit.exe

run: build
	@./bin/bit.exe

test:
	@go test ./... -v