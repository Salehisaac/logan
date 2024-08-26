BINARY_NAME=logan
GO_FLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

build:
	@$(GO_FLAGS) go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd & \
	loader_pid=$$!; \
	{ \
		while kill -0 $$loader_pid 2>/dev/null; do \
			for s in / - \\ \|; do \
				printf "\r\033[1;32m%s\033[0m Building..." "$$s"; \
				sleep 0.1; \
			done; \
		done; \
		printf "\r\033[K"; \
		wait $$loader_pid; \
	}

run: build
	@./$(BINARY_NAME) $(ARGS) 

run-stream: build
	@./$(BINARY_NAME) -s 

clean:
	rm -f $(BINARY_NAME)

.PHONY: build run clean run-stream
