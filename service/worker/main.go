package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	sneaker "github.com/oldfritter/sneaker-go/v3"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/oldfritter/lucy/lib/mq"
	"github.com/oldfritter/lucy/service/worker/sneakerWorker"
)

var (
	closeChan  = make(chan int)
	allWorkers = []sneaker.WorkerI{
		sneakerWorker.CaptchaImageWorkerInstance,
	}
)

func main() {
	initialize()

	go healthCheckLoop()
	startAllWorkers()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")
	go recycle()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	select {
	case <-closeChan:
		cancel()
	case <-ctx.Done():
		cancel()
	}
}

func initialize() {
	setLog()
	setPid()

	// 初始化 RabbitMQ 连接
	mq.InitRabbitMQ()

	// 检查连接是否成功
	conn := mq.GetConnection()
	if conn == nil {
		log.Println("rabbitmq not available, workers will wait for connection")
		return
	}

	// 设置每个 worker 的 RabbitMQ 连接并初始化日志
	for _, w := range allWorkers {
		w.SetRabbitMqConnect(&sneaker.RabbitMqConnect{Connection: conn})
		w.InitLogger()
	}
}

// healthCheckLoop 每 30 秒检查连接状态，断线则重连并重新订阅
func healthCheckLoop() {
	t := time.NewTimer(30 * time.Second)
	for {
		<-t.C
		if mq.IsClosed() {
			if err := mq.Reconnect(); err != nil {
				log.Println("rabbitmq reconnect failed:", err)
				t.Reset(30 * time.Second)
				continue
			}
			log.Println("rabbitmq reconnected")
			startAllWorkers()
		} else {
			// 检查各 worker channel 状态
			for _, w := range allWorkers {
				if w.IsChannelClosed() {
					w.SetRabbitMqConnect(&sneaker.RabbitMqConnect{Connection: mq.GetConnection()})
					sneaker.SubscribeMessageByQueue(w, amqp.Table{})
				}
			}
		}
		t.Reset(30 * time.Second)
	}
}

func startAllWorkers() {
	conn := mq.GetConnection()
	if conn == nil {
		log.Println("rabbitmq not available, skip starting workers")
		return
	}
	for _, w := range allWorkers {
		for i := 0; i < w.GetThreads(); i++ {
			go func(worker sneaker.WorkerI) {
				worker.SetRabbitMqConnect(&sneaker.RabbitMqConnect{Connection: mq.GetConnection()})
				worker.InitLogger()
				sneaker.SubscribeMessageByQueue(worker, amqp.Table{})
			}(w)
		}
	}
}

func setLog() {
	if err := os.Mkdir("logs", 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("logs/workers.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}

func setPid() {
	if err := os.MkdirAll("pids", 0755); err != nil {
		log.Fatalf("create folder error: %v", err)
	}
	if err := os.WriteFile("pids/workers.pid", []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		log.Fatalf("open file error: %v", err)
	}
}

func recycle() {
	for i, w := range allWorkers {
		w.Stop()
		w.Recycle()
		log.Println("stopped:", w.GetName(), "[", i, "]")
	}
	mq.Close()
	closeChan <- 1
}
