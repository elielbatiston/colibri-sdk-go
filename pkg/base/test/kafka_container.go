package test

import (
	"context"
	"fmt"
	"os"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/config"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	kafkaDockerImage  = "confluentinc/cp-kafka:7.9.0"
	kafkaPort         = "9092/tcp"
	testTopicName     = "COLIBRI_PROJECT_USER_CREATE"
	testFailTopicName = "COLIBRI_PROJECT_FAIL_USER_CREATE"
)

var (
	kafkaContainerInstance *KafkaContainer
)

type KafkaContainer struct {
	kafkaContainerRequest *testcontainers.ContainerRequest
	kafkaContainer        testcontainers.Container
	ctx                   context.Context
}

func UseKafkaContainer(ctx context.Context) *KafkaContainer {
	if kafkaContainerInstance == nil {
		kafkaContainerInstance = newKafkaContainer()
		kafkaContainerInstance.ctx = ctx
		kafkaContainerInstance.start()
		kafkaContainerInstance.createTopic()
	}
	return kafkaContainerInstance
}

func newKafkaContainer() *KafkaContainer {
	req := &testcontainers.ContainerRequest{
		Image:        kafkaDockerImage,
		ExposedPorts: []string{"9092:9092"},
		Name:         fmt.Sprintf("colibri-project-test-kafka-%s", uuid.New().String()),
		Env: map[string]string{
			"KAFKA_NODE_ID":                  "1",
			"KAFKA_PROCESS_ROLES":            "broker,controller",
			"KAFKA_CONTROLLER_QUORUM_VOTERS": "1@localhost:9093",

			"KAFKA_LISTENERS":                      "PLAINTEXT://:9092,PLAINTEXT_HOST://:29092,CONTROLLER://:9093",
			"KAFKA_ADVERTISED_LISTENERS":           "PLAINTEXT://localhost:9092",
			"KAFKA_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT,CONTROLLER:PLAINTEXT",

			"KAFKA_INTER_BROKER_LISTENER_NAME": "PLAINTEXT",

			"CLUSTER_ID": "MkU3OEVBNTcwNTJENDM2Qk",

			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
		},
		WaitingFor: wait.ForAll(wait.ForListeningPort(kafkaPort)),
	}

	return &KafkaContainer{kafkaContainerRequest: req}
}

func (c *KafkaContainer) start() {
	var err error
	c.kafkaContainer, err = testcontainers.GenericContainer(c.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *c.kafkaContainerRequest,
		Started:          true,
	})
	if err != nil {
		logging.Fatal(c.ctx).Err(err)
	}

	testDbPort, err := c.kafkaContainer.MappedPort(c.ctx, kafkaPort)
	if err != nil {
		logging.Fatal(c.ctx).Err(err)
	}

	c.setKafkaEnv(testDbPort)

	logging.Info(c.ctx).Msgf("Test Kafka started at port: %s", testDbPort)
}

func (c *KafkaContainer) createTopic() {
	url := os.Getenv(config.ENV_KAFKA_BOOTSTRAP_SERVERS)
	admin, _ := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": url,
	})

	topics := []string{testTopicName, testFailTopicName}
	for _, topic := range topics {
		admin.CreateTopics(
			c.ctx,
			[]kafka.TopicSpecification{
				{
					Topic:             topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			},
		)
		logging.Info(c.ctx).Msgf("Topic %s created", topic)
	}
}

func (c *KafkaContainer) setKafkaEnv(kafkaPort nat.Port) {
	_ = os.Setenv(config.ENV_KAFKA_BOOTSTRAP_SERVERS, fmt.Sprintf("localhost:%s", kafkaPort.Port()))
	_ = os.Setenv(config.ENV_KAFKA_CLIENT_ID, "colibri-sdk-go")
	_ = os.Setenv(config.ENV_KAFKA_CONSUMER_GROUP_ID, "colibri-sdk-go-test")
	_ = os.Setenv(config.ENV_KAFKA_CONSUMER_AUTO_OFFSET_RESET, "earliest")
	_ = os.Setenv(config.ENV_COLIBRI_MESSAGING, "KAFKA")
	config.COLIBRI_MESSAGING = config.MESSAGING_KAFKA
	_ = os.Setenv(config.ENV_CLOUD, config.CLOUD_NONE)
	logging.Info(c.ctx).Msgf("Kafka Bootstrap Server: %s", os.Getenv(config.ENV_KAFKA_BOOTSTRAP_SERVERS))
}
