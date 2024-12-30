package messaging

import (
	"context"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/monitoring"
)

type KafkaProducer struct {
	topic     string
	key       string
	fnSuccess func()
	fnError   func()
}

func (p *KafkaProducer) Publish(ctx context.Context, action string, message any) error {
	return publish(ctx, p, action, message, func() interface{} {
		return monitoring.StartTransactionSegment(ctx, messaging_producer_transaction, map[string]string{
			"topic": p.topic,
			"key":   p.key,
		})
	})
}

func (p *KafkaProducer) GetTopic() string {
	return p.topic
}
