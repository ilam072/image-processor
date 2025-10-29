package consumer

import (
	"context"
	"github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
)

type Consumer struct {
	r *wbfkafka.Consumer
}

func New(brokers []string, topic string, groupId string) *Consumer {
	r := wbfkafka.NewConsumer(
		brokers,
		topic,
		groupId,
	)
	return &Consumer{r: r}
}

func (c *Consumer) Consume(ctx context.Context) (kafka.Message, error) {
	return c.r.Fetch(ctx)
}

func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.r.Commit(ctx, msg)
}

func (c *Consumer) Close() error {
	return c.r.Close()
}
