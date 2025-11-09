package redismq

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

type TestHandler struct {
}

func (t *TestHandler) Process(ctx context.Context, msg *Message) error {
	fmt.Println(fmt.Sprintf("处理消息: %s[%d]", msg.Id, msg.RetryCount))
	return errors.New("test")
}

func (t *TestHandler) Timeout(ctx context.Context, msg *Message) {
	fmt.Println(fmt.Sprintf("处理超时: %s[%d]", msg.Id, msg.RetryCount))
}

func TestConsumer_Start(t *testing.T) {
	stream := ""

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	consumer := NewConsumer(client, stream, &TestHandler{},
		WithMaxIdle(time.Second),
		WithBatchSize(2),
	)
	consumer.Start()
}
