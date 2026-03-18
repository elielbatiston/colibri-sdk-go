package mongoDB

import (
	"context"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/observer"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoDBObserver is a struct for MongoDB observer.
type mongoDBObserver struct {
	instance *mongo.Client
}

// Close finalize MongoDB connection
//
// No parameters.
// No return values.
func (o mongoDBObserver) Close() {
	ctx := context.Background()
	logging.Info(ctx).Msg(dbWaitingSafeClose)

	if observer.WaitRunningTimeout() {
		logging.Warn(ctx).Msg(dbWaitingForceClose)
	}

	if err := o.instance.Disconnect(context.Background()); err != nil {
		logging.Error(ctx).Err(err).Msg(dbCloseError)
	}

	logging.Info(ctx).Msg(dbCloseSuccess)
}
