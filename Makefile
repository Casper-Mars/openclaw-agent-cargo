.PHONY: build clean test install

# 变量
BINARY_NAME = openclaw-agent-cargo
VERSION = 1.0.0
GO = go
GOFLAGS = -ldflags "-X main.version=$(VERSION)"

# 默认目标
all: build

# 构建
build:
	@echo "构建 $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# 跨平台构建
build-all: build-linux build-macos build-windows

build-linux:
	@echo "构建 Linux 版本..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64 .

build-macos:
	@echo "构建 macOS 版本..."
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-macos-arm64 .
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-macos-amd64 .

build-windows:
	@echo "构建 Windows 版本..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# 安装到 GOPATH/bin
install:
	@echo "安装到 GOPATH/bin..."
	$(GO) install $(GOFLAGS) .

# 清理
clean:
	@echo "清理构建文件..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf dist/

# 测试
test:
	@echo "运行测试..."
	$(GO) test -v ./...

# 格式化代码
fmt:
	@echo "格式化代码..."
	$(GO) fmt ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	$(GO) mod download
	$(GO) mod tidy

# 帮助
help:
	@echo "OpenClaw Agent Cargo 构建系统"
	@echo ""
	@echo "可用目标:"
	@echo "  build        - 构建当前平台版本"
	@echo "  build-all    - 构建所有平台版本 (Linux, macOS, Windows)"
	@echo "  install      - 安装到 GOPATH/bin"
	@echo "  clean        - 清理构建文件"
	@echo "  test         - 运行测试"
	@echo "  fmt          - 格式化代码"
	@echo "  deps         - 下载依赖"
	@echo "  help         - 显示此帮助信息"
	@echo ""
