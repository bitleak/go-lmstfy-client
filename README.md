# Go Lmstfy Client ![github actions](https://github.com/bitleak/go-lmstfy-client/actions/workflows/ci.yaml/badge.svg) [![codecov](https://codecov.io/gh/bitleak/go-lmstfy-client/branch/master/graph/badge.svg?token=FXWN0ZN9ZJ)](https://codecov.io/gh/bitleak/go-lmstfy-client) [![Go Report Card](https://goreportcard.com/badge/github.com/bitleak/go-lmstfy-client)](https://goreportcard.com/report/github.com/bitleak/go-lmstfy-client) [![GitHub release date](https://img.shields.io/github/release-date/bitleak/go-lmstfy-client.svg)](https://github.com/bitleak/go-lmstfy-client/releases) [![LICENSE](https://img.shields.io/github/license/bitleak/go-lmstfy-client.svg)](https://github.com/bitleak/go-lmstfy-client/blob/master/LICENSE) [![GoDoc](https://img.shields.io/badge/Godoc-reference-blue.svg)](https://godoc.org/github.com/bitleak/go-lmstfy-client/client)


## Usage

### Setup Lmstfy

As simple as:

```shell
$ make setup
```
it would run the Redis server and lmstfy using docker-compose, and create the new namespace `test-ns` 
with token `01F4CKPFY8XYVH6WVEQB3747CW`.

### Producer Example

```go
import (
	"fmt"
	"github.com/bitleak/go-lmstfy-client/client"
	"os"
)

func main() {

	c := client.NewLmstfyClient("127.0.0.1", 7777, "test-ns", "01F4CKPFY8XYVH6WVEQB3747CW")
	// Optional, config the client to retry when some errors happened. retry 3 times with 50ms interval
	c.ConfigRetry(3, 50)

	// Publish a job with ttl==forever, tries==3, delay==1s
	jobID, err := c.Publish("test-queue", []byte("hello"), 0, 3, 1)
	if err != nil {
		fmt.Printf("Failed to publish the new job, err: %s\n", jobID)
		os.Exit(1)
	}
	fmt.Printf("Published the new job with id: %s\n", jobID)
}
```

[examples/producer](https://github.com/bitleak/go-lmstfy-client/tree/master/examples/producer/main.go) 

### Consumer Example

```go
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

		Threads: 4, // how many polling threads which would affect the performance
		TTR: 30, // unit was second, determine when the job would be visible again
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
```
[examples/consumer](https://github.com/bitleak/go-lmstfy-client/tree/master/examples/consumer/main.go) 

API documentation are available via [godoc](https://godoc.org/github.com/bitleak/go-lmstfy-client/client)
