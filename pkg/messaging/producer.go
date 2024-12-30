package messaging

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/config"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/logging"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/monitoring"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/security"
	"github.com/google/uuid"
)

type ProducerInterface interface {
	Publish(ctx context.Context, action string, message any) error
	GetTopic() string
}

func NewProducer(topicName string) ProducerInterface {
	return &CloudProducer{topicName}
}

func NewKafkaProducer(topicName string, key string, fnSuccess func(), fnError func()) ProducerInterface {
	return &KafkaProducer{topicName, key, fnSuccess, fnError}
}

func publish(ctx context.Context, p ProducerInterface, action string, message any, txnFn func() interface{}) error {
	if instance == nil {
		return errors.New("messaging has not been initialized. add in main.go `messaging.Initialize()`")
	}
	txn := monitoring.GetTransactionInContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			logging.Error("panic recovering publish topic %s: \n%s", p.GetTopic(), string(debug.Stack()))
			monitoring.NoticeError(txn, r.(error))
		}
	}()

	if txn != nil {
		segment := txnFn()
		defer monitoring.EndTransactionSegment(segment)
	}

	msg := &ProviderMessage{
		Id:      uuid.New(),
		Origin:  config.APP_NAME,
		Action:  action,
		Message: message,
	}

	authContext := security.GetAuthenticationContext(ctx)
	if authContext != nil {
		msg.TenantId = authContext.GetTenantID()
		msg.UserId = authContext.GetUserID()
	}

	if err := instance.producer(ctx, p, msg); err != nil {
		logging.Error("Could not send message with id %s to topic %s. Error: %v", msg.Id, p.GetTopic(), err)
		monitoring.NoticeError(txn, err)
		return err
	}

	return nil
}
