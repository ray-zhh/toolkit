package redismq

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

type Consumer struct {
	ctx     context.Context
	client  *redis.Client
	stream  string
	handler Handler
	options *ConsumerOptions
}

type Message struct {
	ID        string
	Payload   string
	RetryCont int
}
type Handler func(message Message) error

type ConsumerOptions struct {
	GroupName     string
	ConsumerName  string
	BatchSize     int64
	BlockDuration time.Duration
}

func NewConsumer(client *redis.Client, stream string, handler Handler, options *ConsumerOptions) *Consumer {
	if options == nil {
		hostname, _ := os.Hostname()
		options = &ConsumerOptions{
			GroupName:     "default-group",
			ConsumerName:  fmt.Sprintf("consumer-%s", hostname),
			BatchSize:     1,
			BlockDuration: time.Second * 5,
		}
	}

	return &Consumer{
		ctx:     context.Background(),
		client:  client,
		stream:  stream,
		handler: handler,
		options: options,
	}
}

func (c *Consumer) Start() error {
	if err := c.initGroup(); err != nil {
		return err
	}
	go c.consume()
	return nil
}

// 初始化消费组
func (c *Consumer) initGroup() error {
	if result, err := c.client.XGroupCreateMkStream(c.ctx, c.stream, c.options.GroupName, "0").Result(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return err
		}
	} else if result != "OK" {
		return fmt.Errorf("non-OK response from Redis: %s", result)
	}
	return nil
}

func (c *Consumer) consume() {
	for {
		message, err := c.client.XReadGroup(c.ctx, &redis.XReadGroupArgs{
			Group:    c.options.GroupName,
			Consumer: c.options.ConsumerName,
			Streams:  []string{c.stream, ">"},
			Count:    c.options.BatchSize,
			Block:    c.options.BlockDuration,
			NoAck:    false,
		}).Result()

		if err != nil {
			continue
		}

		for _, stream := range message {
			for _, msg := range stream.Messages {

			}
		}

	}
}
