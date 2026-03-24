package messaging

import (
	"context"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/config"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/monitoring"
	colibrimonitoringbase "github.com/colibriproject-dev/colibri-sdk-go/pkg/base/monitoring/colibri-monitoring-base"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/security"
	"github.com/google/uuid"
)

type Metadata map[string]string

type publishOptions struct {
	metadata     Metadata
	partitionKey string
}

type PublishOptions func(*publishOptions)

func WithMetadata(md Metadata) PublishOptions {
	return func(o *publishOptions) {
		o.metadata = md
	}
}

func WithPartitionKey(key string) PublishOptions {
	return func(o *publishOptions) {
		o.partitionKey = key
	}
}

type Producer struct {
	topic string
}

func NewProducer(topicName string) *Producer {
	return &Producer{topicName}
}

func (p *Producer) Publish(ctx context.Context, action string, message any, opts ...PublishOptions) error {
	if instance == nil {
		logging.Fatal(context.Background()).Msg(messagingNotInitialized)
	}
	correlationID := ctx.Value(logging.CorrelationIDParam)
	if correlationID == nil {
		correlationID = uuid.New().String()
	}

	txn, _ := monitoring.StartTransaction(ctx, messagingProducerTransaction, colibrimonitoringbase.SpanKindProducer)
	monitoring.AddTransactionAttribute(txn, "topic", p.topic)
	monitoring.AddTransactionAttribute(txn, "correlationId", correlationID.(string))
	monitoring.AddTransactionAttribute(txn, "action", action)
	defer monitoring.EndTransaction(txn)

	msg := &ProviderMessage{
		ID:            uuid.New(),
		Origin:        config.APP_NAME,
		Action:        action,
		Message:       message,
		AuthContext:   security.GetAuthenticationContext(ctx),
		CorrelationID: correlationID.(string),
	}

	options := publishOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if err := instance.producer(ctx, p, msg, options); err != nil {
		logging.Error(ctx).Err(err).Msgf(couldNotSendMsg, msg.ID, p.topic)
		monitoring.NoticeError(txn, err)
		return err
	}

	return nil
}
