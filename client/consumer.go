package client

import (
	"context"
	"fmt"
	"sync"
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

	cfg        *ConsumerConfig
	wg         sync.WaitGroup
	shutdown   chan struct{}
	threadChan chan interface{}
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
		threadChan:   make(chan interface{}, cfg.Threads),
		LmstfyClient: NewLmstfyClient(cfg.Host, cfg.Port, cfg.Namespace, cfg.Token),
	}
}

// Receive would create threads and respawn new goroutine to continue polling jobs when panic.
func (c *Consumer) Receive(ctx context.Context, fn func(ctx context.Context, job *Job)) error {
	for i := 0; i < c.cfg.Threads; i++ {
		c.wg.Add(1)
		c.createThread(ctx, fn)
	}

	go func() {
		for r := range c.threadChan {
			// no need to add semaphore
			c.createThread(ctx, fn)
			fmt.Printf("thread panic, the msg is: %s", r)
		}
	}()

	c.wg.Wait()
	return nil
}

// addThread would create new thread and poll jobs from the remote server.
func (c *Consumer) createThread(ctx context.Context, fn func(ctx context.Context, job *Job)) {
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				// In order to block at Wait(), new thread should add semaphore before panic thread calls Done().
				c.wg.Add(1)
				c.threadChan <- r
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
