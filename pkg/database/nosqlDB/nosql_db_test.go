package nosqlDB

import (
	"context"
	"time"

	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/logging"
	"github.com/colibri-project-io/colibri-sdk-go/pkg/base/test"
)

type Profile struct {
	Id   string `json:"_id" bson:"_id"`
	Name string `json:"name"`
}

type User struct {
	Id       string    `json:"_id" bson:"_id"`
	Name     string    `json:"name"`
	Birthday time.Time `json:"birthday"`
	Profile  Profile   `json:"profile"`
}

func InitializeMongoDBTest(ctx context.Context) {
	basePath := test.MountAbsolutPath(test.NOSQL_ENVIRONMENT_PATH)

	test.InitializeNoSqlDBTest()
	pc := test.UseMongoDBContainer()

	var user User
	datasets := []string{"user1.json", "user2.json"}

	if err := pc.Dataset(user, basePath, datasets...); err != nil {
		logging.Fatal(err.Error())
	}

	Initialize(ctx)
}
