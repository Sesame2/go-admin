MAIN_PATH := ./cmd/server/main.go
APP_NAME := server

.PHONY: swag air clean build air

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