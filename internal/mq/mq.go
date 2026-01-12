package mq

import (
	"context"
)

type Producer interface {
	// 用于发送通用消息
	Send(ctx context.Context, exchange, routingKey string, message []byte) error

	// 健康检查
	HealthCheck() error

	// 关闭连接
	Close() error
}

type Consumer interface {
	// 用于接收通用消息
	Consume(ctx context.Context, queue string, handler func([]byte) error) error

	// 关闭连接
	Close() error
}
