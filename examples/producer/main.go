package main

import (
	"fmt"
	"os"

	"github.com/bitleak/go-lmstfy-client/client"
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
