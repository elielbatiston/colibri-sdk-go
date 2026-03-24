package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

const (
	producerNotInitialized   string = "producer is not initialized"
	consumerNotInitialized   string = "consumer is not initialized"
	couldNotReceiveTopicMsg  string = "error on receive message from topic %s"
	couldNotReadTopicMsgBody string = "could not read message body with id %s from topic %s"

	METADATA_HEADER_KEY            string = "messageKey"
	KAFKA_PREFIX                   string = "KAFKA_"
	KAFKA_PRODUCER_PREFIX          string = "KAFKA_PRODUCER_"
	KAFKA_CONSUMER_PREFIX          string = "KAFKA_CONSUMER_"
	KAFKA_BOOTSTRAP_SERVER_DEFAULT string = "localhost:9092"
)

var (
	ErrFileNotFound = errors.New("file not found")
	ErrReadJsonFile = errors.New("error reading json file")
)

type kafkaMessaging struct {
	kProducer *kafka.Producer
}

type kafkaConsumer struct {
	*kafka.ConfigMap
	*kafka.Consumer
}

type kafkaOriginalMessage struct {
	*kafkaConsumer
	msg *kafka.Message
}

func (k kafkaOriginalMessage) Ack() error {
	ok, err := k.ConfigMap.Get("enable.auto.commit", false)
	if err != nil {
		return err
	}

	if ok == false {
		_, err = k.CommitMessage(k.msg)
		return err
	}

	return nil
}

func (r kafkaOriginalMessage) Nack(_ bool, _ error) error {
	return nil
}

func newKafkaMessaging() *kafkaMessaging {
	return &kafkaMessaging{
		kProducer: newKafkaProducer(),
	}
}

func newKafkaProducer() *kafka.Producer {
	configMap := loadKafkaConfig(KAFKA_PRODUCER_PREFIX, KAFKA_CONSUMER_PREFIX)
	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		logging.Fatal(context.Background()).Err(err).Msg(connectionError)
	}

	return producer
}

func loadKafkaConfig(validPrefix, invalidPrefix string) *kafka.ConfigMap {
	configMap := &kafka.ConfigMap{}
	var prefix string

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := parts[1]

		if strings.Contains(key, invalidPrefix) {
			continue
		}

		if !strings.Contains(key, KAFKA_PREFIX) {
			continue
		}

		prefix = KAFKA_PREFIX
		if strings.HasPrefix(key, validPrefix) {
			prefix = validPrefix
		}

		kafkaKey := strings.TrimPrefix(key, prefix)
		kafkaKey = strings.ToLower(strings.ReplaceAll(kafkaKey, "_", "."))
		configMap.SetKey(kafkaKey, value)
	}

	return configMap
}

func (m *kafkaMessaging) newKafkaConsumer() (*kafkaConsumer, error) {
	configMap := loadKafkaConfig(KAFKA_CONSUMER_PREFIX, KAFKA_PRODUCER_PREFIX)
	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		return nil, err
	}

	return &kafkaConsumer{
		configMap,
		consumer,
	}, nil
}

func (m *kafkaMessaging) producer(ctx context.Context, p *Producer, msg *ProviderMessage, options publishOptions) error {
	if m.kProducer == nil {
		logging.Fatal(ctx).Msg(producerNotInitialized)
	}

	var key string
	if options.partitionKey != "" {
		key = options.partitionKey
	} else {
		key = uuid.New().String()
	}

	var headers []kafka.Header
	for k, v := range options.metadata {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          msg.Byte(),
		Key:            []byte(key),
		Headers:        headers,
	}

	deliveryChan := make(chan kafka.Event, 1)

	err := m.kProducer.Produce(message, deliveryChan)
	if err != nil {
		return err
	}

	go m.deliveryReport(deliveryChan)

	return nil
}

func (m *kafkaMessaging) deliveryReport(deliveryChan chan kafka.Event) {
	for e := range deliveryChan {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logging.Error(context.Background()).Err(ev.TopicPartition.Error).Msgf("failed to deliver message to topic %s", *ev.TopicPartition.Topic)
			}
		}
	}
}

func (m *kafkaMessaging) consumer(ctx context.Context, c *consumer) (chan *ProviderMessage, error) {
	kc, err := m.newKafkaConsumer()
	if err != nil {
		return nil, err
	}

	if kc == nil {
		logging.Fatal(ctx).Msg(consumerNotInitialized)
	}

	err = m.waitTopicReady(ctx, kc.Consumer, c.queue)
	if err != nil {
		return nil, err
	}

	err = kc.Subscribe(c.queue, nil)
	if err != nil {
		return nil, err
	}

	providerMsgs := make(chan *ProviderMessage, 1)

	go m.processMessages(ctx, kc, c, providerMsgs)

	return providerMsgs, nil
}

func (m *kafkaMessaging) processMessages(
	ctx context.Context,
	kc *kafkaConsumer,
	c *consumer,
	providerMsgs chan<- *ProviderMessage,
) {
	for {
		if c.isCanceled() {
			c.Done()
			return
		}

		msg, err := kc.ReadMessage(-1)
		if err != nil {
			logging.Error(ctx).Err(err).Msgf(couldNotReceiveTopicMsg, c.queue)
			continue
		}

		if msg != nil {
			m.handleMessage(ctx, msg, kc, providerMsgs)
		}
	}
}

func (m *kafkaMessaging) handleMessage(ctx context.Context, msg *kafka.Message, kc *kafkaConsumer, providerMsgs chan<- *ProviderMessage) {
	var pm ProviderMessage

	if err := json.Unmarshal(msg.Value, &pm); err != nil {
		logging.Error(ctx).Err(err).Msgf(couldNotReadTopicMsgBody, msg.TopicPartition.Offset, msg.TopicPartition)
	} else {
		pm.addOriginBrokerNotification(kafkaOriginalMessage{kafkaConsumer: kc, msg: msg})
		providerMsgs <- &pm
	}
}

func (m *kafkaMessaging) waitTopicReady(ctx context.Context, c *kafka.Consumer, topic string) error {
	for i := 0; i < 10; i++ {
		meta, err := c.GetMetadata(&topic, false, 1000)
		if err == nil {
			for _, t := range meta.Topics {
				if t.Topic == topic && len(t.Partitions) > 0 {
					logging.Info(ctx).Msgf("topic %s ready", topic)
					return nil
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("topic not ready")
}

func (m *kafkaMessaging) close() error {
	ctx := context.Background()
	logging.Info(ctx).Msg("flushing kafka producer...")

	remaining := m.kProducer.Flush(10000)

	if remaining > 0 {
		logging.Error(ctx).Msgf("failed to deliver %d messages before shutdown", remaining)
	}

	m.kProducer.Close()

	logging.Info(ctx).Msg("producer closed")

	return nil
}
