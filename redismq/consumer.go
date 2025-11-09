package redismq

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"strings"
	"time"
)

const ValuesKey = "message"

type Consumer struct {
	ctx           context.Context
	client        *redis.Client
	stream        string
	group         string
	consumer      string
	batchSize     int64
	handler       Handler
	maxIdle       time.Duration
	maxRetryCount int64
	queue         chan *Message
}

type Handler interface {
	Process(ctx context.Context, message *Message) error
	Timeout(ctx context.Context, message *Message)
}

type Option func(*Consumer)

func WithMaxIdle(maxIdle time.Duration) Option {
	return func(consumer *Consumer) {
		consumer.maxIdle = maxIdle
	}
}
func WithBatchSize(batchSize int64) Option {
	return func(consumer *Consumer) {
		consumer.batchSize = batchSize
	}
}

type Message struct {
	Id         string
	Body       string
	RetryCount int64
}

func NewConsumer(client *redis.Client, stream string, handler Handler, opts ...Option) *Consumer {
	hostname, _ := os.Hostname()
	consumer := &Consumer{
		ctx:           context.Background(),
		client:        client,
		stream:        stream,
		group:         "group_default",
		consumer:      fmt.Sprintf("consumer_%s", strings.ToLower(hostname)),
		batchSize:     100,
		handler:       handler,
		maxIdle:       0,
		maxRetryCount: 0,
		queue:         make(chan *Message),
	}
	for _, opt := range opts {
		opt(consumer)
	}
	return consumer
}

func (c *Consumer) Start() {
	_ = c.xGroupCreateMkStream(c.ctx, c.stream, c.group)
	go c.consume()
	c.produce()
}

// xGroupCreateMkStream 创建消费组
func (c *Consumer) xGroupCreateMkStream(ctx context.Context, stream, group string) error {
	if result, err := c.client.XGroupCreateMkStream(ctx, stream, group, "0").Result(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return err
		}
	} else if result != "OK" {
		return fmt.Errorf("non-OK response from Redis: %s", result)
	}
	return nil
}

func (c *Consumer) produce() {
	for {
		var messages []redis.XMessage
		claims, err := c.claim()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		messages = append(messages, claims...)

		cnt := c.batchSize - int64(len(claims))
		msgs, err := c.xReadGroup(cnt)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		messages = append(messages, msgs...)

		for _, message := range messages {
			retryCount, _ := c.getRetryCount(message.ID)
			fmt.Println(fmt.Sprintf("message: %s, retryCount: %d", message.ID, retryCount))
			c.queue <- &Message{
				Id:         message.ID,
				Body:       message.Values[ValuesKey].(string),
				RetryCount: retryCount,
			}
		}
	}
}

func (c *Consumer) claim() ([]redis.XMessage, error) {
	pendings, err := c.client.XPendingExt(c.ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  "-",
		End:    "+",
		Count:  c.batchSize,
		//Consumer: c.consumer,
	}).Result()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, pending := range pendings {
		if c.maxIdle > 0 && pending.Idle > c.maxIdle {
			ids = append(ids, pending.ID)
			continue
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	messages, err := c.client.XClaim(c.ctx, &redis.XClaimArgs{
		Stream:   c.stream,
		Group:    c.group,
		Consumer: c.consumer,
		MinIdle:  c.maxIdle,
		Messages: ids,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return messages, nil
}

func (c *Consumer) consume() {
	for {
		select {
		case message := <-c.queue:
			go c.handle(message)
		}
	}
}

func (c *Consumer) handle(message *Message) {
	ctx, cancelFunc := context.WithTimeout(c.ctx, c.maxIdle)
	defer cancelFunc()

	if message.RetryCount >= c.maxRetryCount {
		c.handler.Timeout(ctx, message)
	} else if err := c.handler.Process(ctx, message); err != nil {
		fmt.Println(err.Error())
		return
	}
	c.ack(message.Id)
}

func (c *Consumer) ack(id string) {
	pipe := c.client.Pipeline()
	pipe.XAck(c.ctx, c.stream, c.group, id)
	pipe.XDel(c.ctx, c.stream, id)
	_, err := pipe.Exec(c.ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (c *Consumer) xReadGroup(count int64) ([]redis.XMessage, error) {
	xStream, err := c.client.XReadGroup(c.ctx, &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumer,
		Streams:  []string{c.stream, ">"},
		Count:    count,
		Block:    time.Millisecond * 100,
		NoAck:    false,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var messages []redis.XMessage
	for _, message := range xStream {
		messages = append(messages, message.Messages...)
	}
	return messages, nil
}
func (c *Consumer) getRetryCount(id string) (int64, error) {
	info, err := c.client.XPendingExt(c.ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  id,
		End:    id,
		Count:  1,
	}).Result()

	if err != nil {
		return 0, err
	}
	if len(info) == 0 {
		return 0, fmt.Errorf("message not found")
	}

	return info[0].RetryCount, err
}
