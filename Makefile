BINARY_NAME=log_reader
GO_FLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

build:
	$(GO_FLAGS) go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd && upx --brute log_reader

run: build
	@./$(BINARY_NAME)

run-stream: build
	@./$(BINARY_NAME) -s

clean:
	rm -f $(BINARY_NAME)

.PHONY: build run clean run-stream
