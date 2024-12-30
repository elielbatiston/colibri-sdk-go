package messaging

import (
	"context"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/monitoring"
)

type CloudProducer struct {
	topic string
}

func (p *CloudProducer) Publish(ctx context.Context, action string, message any) error {
	return publish(ctx, p, action, message, func() interface{} {
		return monitoring.StartTransactionSegment(ctx, messaging_producer_transaction, map[string]string{
			"topic": p.topic,
		})
	})
}

func (p *CloudProducer) GetTopic() string {
	return p.topic
}
