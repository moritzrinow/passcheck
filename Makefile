ifeq ($(OS),Windows_NT)
	BINARY=passcheck.exe
	GOOS=windows
else
	BINARY=passcheck
	GOOS=linux
endif

all: get build

build:
	go build -o bin/$(BINARY) src/passcheck.go

run:
	go run src/passcheck.go

get:
	go get github.com/syndtr/goleveldb/leveldb
	go get github.com/howeyc/gopass