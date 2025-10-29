package producer

import (
	"context"
	"encoding/json"
	"github.com/ilam072/image-processor/intenal/types/dto"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/wb-go/wbf/kafka"
	"strconv"
)

type Producer struct {
	w *kafka.Producer
}

func New(brokers []string, topic string) *Producer {
	p := kafka.NewProducer(brokers, topic)
	return &Producer{w: p}

}

func (p *Producer) Produce(ctx context.Context, taskID int) error {
	task := dto.TaskMessage{ID: taskID}

	bytes, err := json.Marshal(task)
	if err != nil {
		return errutils.Wrap("failed to marshal task id", err)
	}
	key := []byte(strconv.Itoa(taskID))
	if err := p.w.Send(ctx, key, bytes); err != nil {
		return errutils.Wrap("failed to send taskID to kafka", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.w.Close()
}
