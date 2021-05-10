package client

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

type ConsumerConfig struct {
	Host      string
	Port      int
	Namespace string
	Token     string
	Queues    []string
	TTR       uint32

	Threads int
}

type Consumer struct {
	*LmstfyClient

	ErrorCallback func(err error)

	cfg          *ConsumerConfig
	wg           sync.WaitGroup
	shutdown     chan struct{}
	spawnCounter uint32
}

func (cfg *ConsumerConfig) init() {
	if cfg.Threads <= 0 {
		cfg.Threads = 1
	}
	if cfg.Threads > 64 {
		cfg.Threads = 64
	}
	if cfg.TTR == 0 {
		cfg.TTR = 30
	}
}

// NewConsumer would create a new instance of the high-level consumer
func NewConsumer(cfg *ConsumerConfig) *Consumer {
	cfg.init()
	return &Consumer{
		cfg:          cfg,
		shutdown:     make(chan struct{}),
		LmstfyClient: NewLmstfyClient(cfg.Host, cfg.Port, cfg.Namespace, cfg.Token),
	}
}

// Receive would create threads and wait until all the threads finish.
func (c *Consumer) Receive(ctx context.Context, fn func(ctx context.Context, job *Job)) error {
	for i := 0; i < c.cfg.Threads; i++ {
		c.createThread(ctx, fn)
	}
	c.wg.Wait()
	return nil
}

// createThread would create a new thread and poll jobs from the remote server.
func (c *Consumer) createThread(ctx context.Context, fn func(ctx context.Context, job *Job)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Found panic with err :%s, would respawn a new thread", err)
				atomic.AddUint32(&c.spawnCounter, 1)
				// spawn a new thread when panic
				c.createThread(ctx, fn)
			}
		}()

		for {
			select {
			case <-c.shutdown:
				return
			default:
			}

			job, err := c.ConsumeFromQueues(c.cfg.TTR, 3, c.cfg.Queues...)
			// err == nil and job == nil means no job
			if err == nil && job == nil {
				continue
			}
			if err != nil {
				if c.ErrorCallback != nil {
					c.ErrorCallback(err)
				}
				continue
			}
			fn(ctx, job)
		}
	}()
}

// Close would stop all polling threads
func (c *Consumer) Close() {
	close(c.shutdown)
}
