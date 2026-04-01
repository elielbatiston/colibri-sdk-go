package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/config"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mongoDBDockerImage  = "mongo:5.0"
	testMongoDBHost     = "localhost:%s"
	testMongoDBName     = "test_db"
	testMongoDBUser     = "root"
	testMongoDBPassword = "root"
	testMongoDBSvcPort  = "27017/tcp"
	testMongoDBDSN      = "mongodb://%s:%s@%s/%s?authSource=admin"
	json_parse_error    = "error when parsing data: %s"
)

var (
	ErrFileNotFound          = errors.New("file not found")
	ErrReadJsonFile          = errors.New("error reading json file")
	mongoDBContainerInstance *MongoDBContainer
)

type MongoDBContainer struct {
	mongoDBContainerRequest *testcontainers.ContainerRequest
	mongoDBContainer        testcontainers.Container
	mongoDBClient           *mongo.Client
	ctx                     context.Context
}

// UseMongoDBContainer initialize container for integration tests.
func UseMongoDBContainer(ctx context.Context) *MongoDBContainer {
	if mongoDBContainerInstance == nil {
		mongoDBContainerInstance = newMongoDBContainer()
		mongoDBContainerInstance.ctx = ctx
		mongoDBContainerInstance.start()
	}
	return mongoDBContainerInstance
}

func newMongoDBContainer() *MongoDBContainer {
	req := &testcontainers.ContainerRequest{
		Image:        mongoDBDockerImage,
		ExposedPorts: []string{testMongoDBSvcPort},
		Name:         fmt.Sprintf("colibri-project-test-mongodb-%s", uuid.New().String()),
		Env: map[string]string{
			"MONGO_INITDB_DATABASE":      testMongoDBName,
			"MONGO_INITDB_ROOT_USERNAME": testMongoDBUser,
			"MONGO_INITDB_ROOT_PASSWORD": testMongoDBPassword,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(testMongoDBSvcPort),
		),
	}

	return &MongoDBContainer{mongoDBContainerRequest: req}
}

func (c *MongoDBContainer) start() {
	var err error
	c.mongoDBContainer, err = testcontainers.GenericContainer(mongoDBContainerInstance.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *c.mongoDBContainerRequest,
		Started:          true,
	})
	if err != nil {
		logging.Fatal(mongoDBContainerInstance.ctx).Err(err)
	}

	testDbPort, err := c.mongoDBContainer.MappedPort(mongoDBContainerInstance.ctx, testMongoDBSvcPort)
	if err != nil {
		logging.Fatal(mongoDBContainerInstance.ctx).Err(err)
	}

	logging.Info(mongoDBContainerInstance.ctx).Msgf("Test mongodb started at port: %s", testDbPort)

	c.setDatabaseEnv(testDbPort)

	dsn := os.Getenv(config.ENV_MONGODB_URI)

	clientOptions := options.Client().ApplyURI(dsn)
	if c.mongoDBClient, err = mongo.Connect(mongoDBContainerInstance.ctx, clientOptions); err != nil {
		logging.Fatal(mongoDBContainerInstance.ctx).Err(err)
	}
}

func (c *MongoDBContainer) Dataset(model any, collectionName string, basePath string, scripts ...string) error {
	collection := c.mongoDBClient.Database(testMongoDBName).Collection(collectionName)

	collection.DeleteMany(mongoDBContainerInstance.ctx, bson.D{})

	for _, s := range scripts {
		filePath := fmt.Sprintf("%s/%s", basePath, s)

		err := Parse(filePath, &model)
		if err != nil {
			return err
		}

		_, err = collection.InsertOne(mongoDBContainerInstance.ctx, model)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *MongoDBContainer) DeleteAll(collectionName string, model any) {
	collection := c.mongoDBClient.Database(testMongoDBName).Collection(collectionName)
	collection.DeleteMany(mongoDBContainerInstance.ctx, bson.D{})
}

func (c *MongoDBContainer) setDatabaseEnv(testDbPort nat.Port) {
	port := strings.Replace(testDbPort.Port(), "/tcp", "", 1)
	port = strings.TrimSpace(port)

	hosts := fmt.Sprintf(testMongoDBHost, port)
	dsn := fmt.Sprintf(testMongoDBDSN,
		testMongoDBUser,
		testMongoDBPassword,
		hosts,
		testMongoDBName,
	)

	c.setEnv(config.ENV_MONGODB_URI, dsn)
}

func (c *MongoDBContainer) setEnv(env string, value string) {
	if err := os.Setenv(env, value); err != nil {
		logging.Fatal(mongoDBContainerInstance.ctx).Msgf("could not set env[%s] value[%s]: %v", env, value, err)
	}
}

func Load(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, ErrReadJsonFile
	}

	return jsonData, nil
}

func Parse(filePath string, model interface{}) error {
	jsonData, err := Load(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, model)
	if err != nil {
		return fmt.Errorf(json_parse_error, err)
	}

	return nil
}
