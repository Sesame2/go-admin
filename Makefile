MAIN_PATH := ./cmd/server/main.go
APP_NAME := server

.PHONY: swag air clean build run test test-coverage

# 生成swagger文档
swag:
	@echo "生成swagger文档"
	swag init -g $(MAIN_PATH)

air:
	@echo "使用air热重载启动服务（开发模式）"
	air

clean:
	@echo "清理构建产物"
	rm -rf bin/*

build:
	@echo "开始构建应用"
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run:
	@echo "启动应用"
	./bin/$(APP_NAME)

# 运行所有测试
test:
	@echo "运行测试"
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行特定包的测试
test-pkg:
	@echo "运行指定包的测试 (使用 PKG=<package-path>)"
	go test -v $(PKG)