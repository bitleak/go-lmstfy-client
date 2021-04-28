package client

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsumer_Receive(t *testing.T) {
	const testConsumeQueue = "test_consumer_queue"

	cli := NewLmstfyClient(TestHost, TestPort, TestNamespace, TestToken)

	var jobIDs []string
	for i := 0; i < 3; i++ {
		jobID, err := cli.Publish(testConsumeQueue, []byte(strconv.Itoa(i)), 0, 1, 0)
		require.Nil(t, err)
		jobIDs = append(jobIDs, jobID)
	}

	var ind int
	consumer := NewConsumer(&ConsumerConfig{
		Host:      TestHost,
		Port:      TestPort,
		Namespace: TestNamespace,
		Token:     TestToken,
		Queues:    []string{testConsumeQueue},
	})
	err := consumer.Receive(context.Background(), func(ctx context.Context, job *Job) {
		if jobIDs[ind] != job.ID {
			require.Equal(t, jobIDs[ind], job.ID, "Mismatched job id")
		}
		require.Nil(t, job.Ack())
		ind++
		if ind == len(jobIDs) {
			consumer.Close()
		}
	})
	require.Nil(t, err)
}
