.PHONY: all

all:
	GOOS=linux GOARCH=amd64 go build
