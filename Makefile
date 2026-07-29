BINARY_NAME := md2jira

.PHONY: build clean

build:
	go build -o $(BINARY_NAME) ./cmd

clean:
	rm -f $(BINARY_NAME)