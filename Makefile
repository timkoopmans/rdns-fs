.PHONY: build clean

build:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/rdns-fs -a -ldflags '-s -w -extldflags "-static"' main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock
