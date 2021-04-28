package client

import (
	"context"
	"os"
	"strconv"
	"testing"
)

func TestConsumer_Receive(t *testing.T) {
	const testConsumeQueue = "test-comsume"

	cli := NewLmstfyClient(Host, Port, Namespace, Token)
	consumer := NewConsumer(ConsumerConfig{
		Host:          Host,
		Port:          Port,
		Namespace:     Namespace,
		Token:         Token,
		Queue:         testConsumeQueue,
		TimeoutSecond: 3,
		TickSecond:    1,
	})

	var jobIDList []string
	for i := 0; i < 3; i++ {
		jobID, _ := cli.Publish(testConsumeQueue, []byte(strconv.Itoa(i)), 0, 1, 0)
		jobIDList = append(jobIDList, jobID)
	}

	var count int
	consumer.Receive(context.Background(), func(ctx context.Context, job *Job) error {
		if jobIDList[count] != job.ID {
			t.Errorf("Mismatched job id, the expected is %s, the real is %s", jobIDList[count], job.ID)
		}
		consumer.Ack(job)

		count++
		if count == len(jobIDList) {
			os.Exit(0)
		}
		return nil
	})
}
