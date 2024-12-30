package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/config"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/logging"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type kafkaMessaging struct {
	kProducer *ckafka.Producer
	kConsumer *ckafka.Consumer
}

func newKafkaMessaging() *kafkaMessaging {
	kMessaging := &kafkaMessaging{
		kProducer: newKafkaProducer(),
		kConsumer: newKafkaConsumer(),
	}

	return kMessaging
}

func newKafkaProducer() *ckafka.Producer {
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": os.Getenv(config.ENV_KAFKA_BOOTSTRAP_SERVER),
		"client.id":         os.Getenv(config.ENV_KAFKA_CLIENT_ID),
	}

	for key, value := range config.KAFKA_PRODUCER_OPTIONS {
		configMap.SetKey(key, value)
	}

	producer, err := ckafka.NewProducer(configMap)
	if err != nil {
		logging.Fatal(connection_error, err)
	}

	return producer
}

func newKafkaConsumer() *ckafka.Consumer {
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": os.Getenv(config.ENV_KAFKA_BOOTSTRAP_SERVER),
		"client.id":         os.Getenv(config.ENV_KAFKA_CLIENT_ID),
		"group.id":          os.Getenv(config.ENV_KAFKA_CONSUMER_GROUP_ID),
	}

	for key, value := range config.KAFKA_CONSUMER_OPTIONS {
		configMap.SetKey(key, value)
	}

	consumer, err := ckafka.NewConsumer(configMap)
	if err != nil {
		logging.Fatal(connection_error, err)
	}

	return consumer
}

func (m *kafkaMessaging) producer(ctx context.Context, p ProducerInterface, msg *ProviderMessage) error {
	deliveryChan := make(chan ckafka.Event)
	pKafka, ok := p.(*KafkaProducer)
	if !ok {
		return errors.New("parameter p is not of type *ProducerKafka")
	}

	message := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &pKafka.topic, Partition: ckafka.PartitionAny},
		Value:          msg.Byte(),
		Key:            []byte(pKafka.key),
	}

	go m.deliveryReport(pKafka, deliveryChan)

	return m.kProducer.Produce(message, deliveryChan)
}

func (m *kafkaMessaging) deliveryReport(p *KafkaProducer, deliveryChan chan ckafka.Event) {
	for e := range deliveryChan {
		switch ev := e.(type) {
		case *ckafka.Message:
			if ev.TopicPartition.Error != nil {
				if p.fnError != nil {
					p.fnError()
				}
			} else {
				if p.fnSuccess != nil {
					p.fnSuccess()
				}
			}
		}
	}
}

func (m *kafkaMessaging) consumer(ctx context.Context, c *consumer) (chan *ProviderMessage, error) {
	ch := make(chan *ProviderMessage, 1)
	err := m.kConsumer.SubscribeTopics([]string{c.queue}, nil)
	if err != nil {
		logging.Fatal("Failed to subscribe to topic: %v", err)
	}

	go func() {
		for {
			if c.isCanceled() {
				c.Done()
				return
			}
			msg, err := m.kConsumer.ReadMessage(-1)
			if err != nil {
				logging.Error("Kafka: Could not read messages from queue %s. Error: %v", c.queue, err)
				continue
			}

			var pm ProviderMessage
			if err = json.Unmarshal(msg.Value, &pm); err != nil {
				logging.Error(couldNotReadMsgBody, msg.TopicPartition.Offset, c.queue, err)
			} else {
				ch <- &pm
				if _, err := m.kConsumer.CommitMessage(msg); err != nil {
					logging.Error("Error confirming message: %v", err)
				}
			}
		}
	}()

	return ch, nil
}
