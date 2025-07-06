package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/Sesame2/go-admin/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type RabbitMQClient struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	logger      *zap.Logger
	config      *config.RabbitMQConfig
	isConnected bool
}

func NewRabbitMQClient(config *config.RabbitMQConfig, logger *zap.Logger) *RabbitMQClient {
	if config.ExchangeType == "" {
		config.ExchangeType = "topic"
	}
	if config.ExchangeName == "" {
		config.ExchangeName = "go_admin_exchange"
	}

	return &RabbitMQClient{
		config: config,
		logger: logger,
	}
}

func (c *RabbitMQClient) Connect() error {
	if c.isConnected {
		return nil
	}
	c.logger.Info("连接到 RabbitMQ",
		zap.String("host", c.config.Host),
		zap.Int("port", c.config.Port))
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.config.Username,
		c.config.Password,
		c.config.Host,
		c.config.Port,
		c.config.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		c.logger.Error("连接Rabbitmq失败", zap.Error(err))
		return fmt.Errorf("连接Rabbitmq失败:%w", err)
	}
	c.conn = conn

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		c.logger.Error("创建通道失败", zap.Error(err))
		return fmt.Errorf("创建通道失败：%w", err)
	}
	c.channel = ch

	// 交换机声明
	err = ch.ExchangeDeclare(
		c.config.ExchangeName,
		c.config.ExchangeType,
		c.config.Durable,
		false, // 自动删除
		false, // 内部的
		false, // no-wait
		nil,   // 其他参数
	)
	if err != nil {
		ch.Close()
		conn.Close()
		c.logger.Error("声明交换机失败,", zap.Error(err))
		return fmt.Errorf("声明交换机失败：%w", err)
	}

	c.isConnected = true
	c.logger.Info("RabbitMQ连接成功")

	// 设置连接关闭通知
	go func() {
		<-conn.NotifyClose(make(chan *amqp.Error))
		c.logger.Warn("Rabbitmq 连接已关闭")
		c.isConnected = false
	}()
	return nil
}

func (c *RabbitMQClient) Send(ctx context.Context, exchange, routingKey string, message []byte) error {
	if !c.isConnected {
		if err := c.Connect(); err != nil {
			return err
		}
	}
	if exchange == "" {
		exchange = c.config.ExchangeName
	}

	c.logger.Debug(
		"发送消息",
		zap.String("exchange", exchange),
		zap.String("routing_key", routingKey),
	)

	return c.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent, // 持久化
			Timestamp:    time.Now(),
		},
	)
}
