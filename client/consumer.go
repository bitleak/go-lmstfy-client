package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ConsumerConfig struct {
	Host          string
	Port          int
	Namespace     string
	Token         string
	Queue         string
	TimeoutSecond uint32
	TickSecond    uint32
}

type Consumer struct {
	*LmstfyClient
	Queue         string // target queue
	TimeoutSecond uint32 // timeout from poll job from queue
	TickSecond    uint32 // time cycle for poll job from queue
}

func NewConsumer(cc ConsumerConfig) *Consumer {
	return &Consumer{
		LmstfyClient:  NewLmstfyClient(cc.Host, cc.Port, cc.Namespace, cc.Token),
		Queue:         cc.Queue,
		TimeoutSecond: cc.TimeoutSecond,
		TickSecond:    cc.TickSecond,
	}
}

// registerSignal register specified signals that could stop blocking consumer receive function
func registerSignal(shutdown chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1}...)
	go func() {
		for range c {
			fmt.Println("Receive shutdown signal...")
			close(shutdown)
			return
		}
	}()
}

// Receive receive job from queue continuously and stop when specified signals are sent
func (c *Consumer) Receive(ctx context.Context, fn func(ctx context.Context, job *Job) error) {
	shutdown := make(chan struct{})
	registerSignal(shutdown)

	timer := time.NewTicker(time.Duration(c.TickSecond) * time.Second)
	defer timer.Stop()
	go func() {
		for {
			<-timer.C
			job, err := c.Consume(c.Queue, 10, c.TimeoutSecond)
			if err != nil || job == nil {
				continue
			}
			err = fn(ctx, job)
			if err != nil {
				continue
			}
		}
	}()

	<-shutdown
	return
}

// Ack delete job from queue
func (c *Consumer) Ack(job *Job) error {
	return c.LmstfyClient.Ack(job.Queue, job.ID)
}
