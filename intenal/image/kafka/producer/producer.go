package producer

import (
	"context"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/wb-go/wbf/kafka"
)

type Producer struct {
	w *kafka.Producer
}

func New(brokers []string, topic string) *Producer {
	p := kafka.NewProducer(brokers, topic)
	return &Producer{w: p}

}

func (p *Producer) Produce(ctx context.Context, taskID string) error {
	if err := p.w.Send(ctx, nil, []byte(taskID)); err != nil {
		return errutils.Wrap("failed to send taskID to kafka", err)
	}

	return nil
}
