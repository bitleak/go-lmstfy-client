package client

import (
	"context"
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

	cfg      *ConsumerConfig
	wg       sync.WaitGroup
	shutdown chan struct{}
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

// Receive would create threads and poll jobs from the remote server
func (c *Consumer) Receive(ctx context.Context, fn func(ctx context.Context, job *Job)) error {
	c.wg.Add(c.cfg.Threads)
	for i := 0; i < c.cfg.Threads; i++ {
		go func() {
			defer c.wg.Done()

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

	c.wg.Wait()
	return nil
}

// Close would stop the poll thread
func (c *Consumer) Close() {
	close(c.shutdown)
}
