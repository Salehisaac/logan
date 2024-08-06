
BINARY_NAME=log_reader



build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .


run: build
	@./$(BINARY_NAME) 


run-stream: build
	@./$(BINARY_NAME) -s


clean:
	rm -f $(BINARY_NAME)


.PHONY: build run clean run-stream
