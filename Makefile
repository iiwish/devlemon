.PHONY: all build app run-app clean test

all: build app

build:
	go build -ldflags "-s -w" -o devlemon ./cmd/devlemon
	ln -sf devlemon dl

app:
	bash ./mac/scripts/build_app.sh

run-app: app
	open ./build/DevLemon.app

test:
	go test ./...

clean:
	rm -rf devlemon dl build .build mac/.build
