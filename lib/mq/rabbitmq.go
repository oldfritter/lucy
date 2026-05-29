package mq

import (
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/oldfritter/lucy/base"
)

var (
	conn     *amqp.Connection
	connOnce sync.Once
	connMu   sync.RWMutex
)

// RabbitConfig RabbitMQ 连接配置
type RabbitConfig struct {
	URL string
}

// buildURL 根据配置拼接 AMQP URL；优先 rabbitmq.url，否则用独立字段
func buildURL() string {
	config := base.GetConfig()

	if url := config.Get("rabbitmq.url", ""); url != "" {
		return url
	}

	username := config.Get("rabbitmq.username", "guest")
	password := config.Get("rabbitmq.password", "guest")
	host := config.Get("rabbitmq.host", "localhost")
	port := config.Get("rabbitmq.port", "5672")
	vhost := config.Get("rabbitmq.vhost", "/")

	return "amqp://" + username + ":" + password + "@" + host + ":" + port + "/" + vhost
}

// InitRabbitMQ 初始化 RabbitMQ 连接（支持多次调用，实际只初始化一次）
func InitRabbitMQ() {
	connOnce.Do(func() {
		initConn(buildURL())
	})
}

// InitRabbitMQWithURL 使用指定 URL 初始化（覆盖配置，只初始化一次）
func InitRabbitMQWithURL(url string) {
	connOnce.Do(func() {
		initConn(url)
	})
}

func initConn(url string) {
	var err error
	connMu.Lock()
	defer connMu.Unlock()

	conn, err = amqp.Dial(url)
	if err != nil {
		log.Printf("rabbitmq dial failed: %v", err)
		return
	}

	log.Printf("rabbitmq connected")
	fmt.Println("rabbitmq connected")
	go func() {
		reason, ok := <-conn.NotifyClose(make(chan *amqp.Error))
		if ok {
			log.Printf("rabbitmq connection closed: %v", reason)
		}
	}()
}

// GetConnection 返回当前 RabbitMQ 连接；未初始化则自动初始化
func GetConnection() *amqp.Connection {
	if conn == nil {
		InitRabbitMQ()
	}
	connMu.RLock()
	defer connMu.RUnlock()
	return conn
}

// GetChannel 从当前连接创建新 Channel
func GetChannel() (*amqp.Channel, error) {
	c := GetConnection()
	if c == nil || c.IsClosed() {
		return nil, amqp.ErrClosed
	}
	return c.Channel()
}

// IsClosed 判断连接是否已关闭
func IsClosed() bool {
	connMu.RLock()
	defer connMu.RUnlock()
	if conn == nil {
		return true
	}
	return conn.IsClosed()
}

// Reconnect 重连 RabbitMQ（关闭旧连接后重新 Dial）
func Reconnect() error {
	connMu.Lock()
	defer connMu.Unlock()

	if conn != nil && !conn.IsClosed() {
		if err := conn.Close(); err != nil {
			log.Printf("rabbitmq close old connection error: %v", err)
		}
	}

	var err error
	conn, err = amqp.Dial(buildURL())
	if err != nil {
		log.Printf("rabbitmq reconnect failed: %v", err)
		return err
	}
	log.Printf("rabbitmq reconnected")
	fmt.Println("rabbitmq reconnected")
	return nil
}

// Close 关闭 RabbitMQ 连接
func Close() {
	connMu.Lock()
	defer connMu.Unlock()

	if conn != nil && !conn.IsClosed() {
		if err := conn.Close(); err != nil {
			log.Printf("rabbitmq close error: %v", err)
		}
	}
}

// DeclareQueue 声明队列（durable 持久化）
func DeclareQueue(name string) (amqp.Queue, error) {
	ch, err := GetChannel()
	if err != nil {
		return amqp.Queue{}, err
	}
	defer ch.Close()

	return ch.QueueDeclare(name, true, false, false, false, nil)
}

// PublishMessage 发送消息到指定队列
func PublishMessage(queue string, body []byte) error {
	ch, err := GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

// ConsumeMessage 消费队列消息
func ConsumeMessage(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	ch, err := GetChannel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}

	return ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}
