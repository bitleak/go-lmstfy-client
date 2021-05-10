package client

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsumer_Receive(t *testing.T) {
	const testConsumerQueue = "test_consumer_queue"

	cli := NewLmstfyClient(TestHost, TestPort, TestNamespace, TestToken)

	var jobIDMap sync.Map
	for i := 0; i < 3; i++ {
		jobID, err := cli.Publish(testConsumerQueue, []byte(strconv.Itoa(i)), 0, 1, 0)
		require.Nil(t, err)
		jobIDMap.Store(jobID, false)
	}

	consumer := NewConsumer(&ConsumerConfig{
		Host:      TestHost,
		Port:      TestPort,
		Namespace: TestNamespace,
		Token:     TestToken,
		Queues:    []string{testConsumerQueue},
	})
	err := consumer.Receive(context.Background(), func(ctx context.Context, job *Job) {
		isConsumed, ok := jobIDMap.Load(job.ID)
		if !ok || isConsumed.(bool) {
			require.Equal(t, false, isConsumed.(bool), "The job should not be consumed.")
		}
		require.Nil(t, job.Ack())

		jobIDMap.Store(job.ID, true)

		isAllConsumed := true
		jobIDMap.Range(func(key, value interface{}) bool {
			if !value.(bool) {
				isAllConsumed = false
				return false
			}
			return true
		})
		if isAllConsumed {
			consumer.Close()
		}
	})
	require.Nil(t, err)
}

func TestConsumer_Respawn(t *testing.T) {
	const testConsumerQueue = "test_consumer_queue"

	cli := NewLmstfyClient(TestHost, TestPort, TestNamespace, TestToken)

	var jobIDMap sync.Map
	for i := 0; i < 10; i++ {
		jobID, err := cli.Publish(testConsumerQueue, []byte(strconv.Itoa(i)), 0, 1, 0)
		require.Nil(t, err)
		jobIDMap.Store(jobID, false)
	}

	consumer := NewConsumer(&ConsumerConfig{
		Host:      TestHost,
		Port:      TestPort,
		Namespace: TestNamespace,
		Token:     TestToken,
		Queues:    []string{testConsumerQueue},
	})
	var flag int32
	err := consumer.Receive(context.Background(), func(ctx context.Context, job *Job) {
		isConsumed, ok := jobIDMap.Load(job.ID)
		if !ok || isConsumed.(bool) {
			require.Equal(t, false, isConsumed.(bool), "The job should not be consumed.")
		}
		require.Nil(t, job.Ack())

		jobIDMap.Store(job.ID, true)

		isAllConsumed := true
		jobIDMap.Range(func(key, value interface{}) bool {
			if !value.(bool) {
				isAllConsumed = false
				return false
			}
			return true
		})
		if isAllConsumed {
			consumer.Close()
		}

		if ok := atomic.CompareAndSwapInt32(&flag, 0, 1); ok {
			panic("Some errors occur!")
		}
	})
	require.Nil(t, err)

	require.Equal(t, uint32(1), consumer.spawnCounter, "The spawn counter should be 1.")
}
