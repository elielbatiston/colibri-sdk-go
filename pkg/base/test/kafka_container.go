package test

import (
	"context"
	"os"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/config"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/logging"

	testcontainerkafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const kafkaDockerImage = "confluentinc/confluent-local:7.5.0"

var kafkaContainerInstance *KafkaContainer

type KafkaContainer struct {
	kafkaContainer *testcontainerkafka.KafkaContainer
}

// UseKafkaContainer initialize container for integration tests.
func UseKafkaContainer() *KafkaContainer {
	if kafkaContainerInstance == nil {
		kafkaContainerInstance = newKafkaContainer()
	}
	return kafkaContainerInstance
}

func newKafkaContainer() *KafkaContainer {
	ctx := context.Background()

	testContanerKafkaContainer, err := testcontainerkafka.Run(ctx,
		kafkaDockerImage,
	)
	if err != nil {
		logging.Fatal(err.Error())
	}

	brokerAddress, err := testContanerKafkaContainer.Brokers(ctx)
	if err != nil {
		logging.Fatal(err.Error())
	}

	logging.Info("Test kafka started at port: %s", brokerAddress[0])

	kafkaContainer := &KafkaContainer{kafkaContainer: testContanerKafkaContainer}

	kafkaContainer.setDatabaseEnv(brokerAddress[0])

	return kafkaContainer
}

func (c *KafkaContainer) setDatabaseEnv(testKafkaPort string) {
	c.setEnv(config.ENV_KAFKA_BOOTSTRAP_SERVER, testKafkaPort)
}

func (c *KafkaContainer) setEnv(env string, value string) {
	if err := os.Setenv(env, value); err != nil {
		logging.Fatal("could not set env[%s] value[%s]: %v", env, value, err)
	}
}
