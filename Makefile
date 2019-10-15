GOOS=windows

ifeq ($(OS),Windows_NT)
	BINARY=passcheck.exe
else
	BINARY=passcheck
endif

all: get build

build:
	go build -o bin/$(BINARY) src/passcheck.go

run:
	go run src/passcheck.go

get:
	go get github.com/syndtr/goleveldb/leveldb
	go get github.com/howeyc/gopass