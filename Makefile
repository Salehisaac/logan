
BINARY_NAME=log_reader



build:
	go build -o $(BINARY_NAME) .


run: build
	@./$(BINARY_NAME) 


run-stream: build
	@./$(BINARY_NAME) -s


clean:
	rm -f $(BINARY_NAME)


.PHONY: build run clean run-stream
