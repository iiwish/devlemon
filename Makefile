.PHONY: all build app app-direct app-mas pkg-mas run-app clean test

all: build app

build:
	go build -ldflags "-s -w" -o devlemon ./cmd/devlemon
	ln -sf devlemon dl

# 默认构建：GitHub 直装版 (Sparkle 2.0 自动更新 + CLI 自动安装)
app: app-direct

app-direct:
	bash ./mac/scripts/build_app.sh direct

# Mac App Store 规范构建 (开启沙盒 + 目录授权 + 移除 Sparkle)
app-mas:
	bash ./mac/scripts/build_app.sh mas

# Mac App Store 签名封装 (.pkg 上架包)
pkg-mas:
	bash ./mac/scripts/package_mas.sh

run-app: app
	open ./build/DevLemon.app

test:
	go test ./...

clean:
	rm -rf devlemon dl build .build mac/.build
