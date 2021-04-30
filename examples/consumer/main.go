package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitleak/go-lmstfy-client/client"
)

func registerSignal(shutdownFn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	}...)
	go func() {
		for range c {
			shutdownFn()
			os.Exit(0)
		}
	}()
}

func main() {
	consumer := client.NewConsumer(&client.ConsumerConfig{
		Host:      "127.0.0.1",
		Port:      7777,
		Namespace: "test-ns",
		Token:     "01F4CKPFY8XYVH6WVEQB3747CW",
		Queues:    []string{"test-queue"},

		Threads: 4,  // how many polling threads which would affect the performance
		TTR:     30, // unit was second, determine when the job would be visible again
	})

	registerSignal(func() {
		consumer.Close()
		fmt.Println("Bye, Going to shutdown the consumer")
	})

	fmt.Println("Hello, Going to consume jobs from queue")
	_ = consumer.Receive(context.Background(), func(ctx context.Context, job *client.Job) {
		fmt.Printf("Got new job: %s\n", job.ID)
		if err := job.Ack(); err != nil {
			fmt.Printf("Failed to ack the job: %s, err: %s\n", job.ID, err.Error())
		}
	})
}
