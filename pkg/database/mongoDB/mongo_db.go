package mongoDB

import (
	"context"
	"time"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/config"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/observer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const (
	dbConnectionSuccess   string = "MongoDB connected"
	dbAlreadyConnected    string = "MongoDB already connected"
	dbConnectionError     string = "An error occurred while trying to connect to the MongoDB. Error: %s"
	dbWaitingSafeClose    string = "waiting to safely close the MongoDB connection"
	dbWaitingForceClose   string = "waiting timed out, forcing to close the MongoDB connection"
	dbCloseError          string = "error on closing the MongoDB connection"
	dbCloseSuccess        string = "MongoDB closed"
	dbNotInitializedError string = "database not initialized"
	modelIsEmptyError     string = "model is empty"
	pageIsEmptyError      string = "page is empty"
)

// mongoDBInstance is a pointer to mongo.Client
var mongoDBInstance *mongo.Client

// Initialize start connection with mongo database and execute migration.
//
// No parameters.
// No return values.
func Initialize() {
	if mongoDBInstance != nil {
		logging.Info(context.Background()).Msg(dbAlreadyConnected)
		return
	}

	mongoDB := NewMongoDBInstance(config.MONGODB_CONNECTION_URI)
	mongoDBInstance = mongoDB
}

// NewMongoDBInstance creates a new MongoDB instance.
//
// Parameters:
// - dsn: a string representing the connection string.
// Returns a pointer to mongo.Client.
func NewMongoDBInstance(dsn string) *mongo.Client {
	mongodbMaxConnIdleTime := time.Duration(config.MONGODB_MAX_CONN_IDLE_TIME) * time.Minute
	mongodbConnectionTimeout := time.Duration(config.MONGODB_CONNECT_TIMEOUT) * time.Millisecond
	mongodbServerSelectionTimeout := time.Duration(config.MONGODB_SERVER_SELECTION_TIMEOUT) * time.Millisecond
	clientOptions := options.Client().
		ApplyURI(dsn).
		SetMinPoolSize(uint64(config.MONGODB_MIN_POOL_SIZE)).
		SetMaxPoolSize(uint64(config.MONGODB_MAX_POOL_SIZE)).
		SetMaxConnIdleTime(mongodbMaxConnIdleTime).
		SetConnectTimeout(mongodbConnectionTimeout).
		SetServerSelectionTimeout(mongodbServerSelectionTimeout).
		SetMonitor(otelmongo.NewMonitor())

	mongoDB, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logging.Fatal(context.Background()).Msgf(dbConnectionError, err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), mongodbServerSelectionTimeout)
	defer cancel()

	if err = mongoDB.Ping(pingCtx, getReadPref()); err != nil {
		logging.Fatal(context.Background()).Msgf(dbConnectionError, err)
	}

	logging.Info(context.Background()).Msgf(dbConnectionSuccess)
	observer.Attach(mongoDBObserver{mongoDB})

	return mongoDB
}
