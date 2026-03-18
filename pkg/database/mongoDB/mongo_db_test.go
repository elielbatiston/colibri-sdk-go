package mongoDB

import (
	"context"
	"time"

	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/logging"
	"github.com/colibriproject-dev/colibri-sdk-go/pkg/base/test"
)

type Profile struct {
	Id   string `json:"_id" bson:"_id"`
	Name string `json:"name"`
}

type User struct {
	ID       string    `json:"_id" bson:"_id"`
	Name     string    `json:"name"`
	Birthday time.Time `json:"birthday"`
	Profile  Profile   `json:"profile"`
}

func (u User) CollectionName() string {
	return "user"
}

func (u User) GetID() string {
	return u.ID
}

func InitializeMongoDBTest(ctx context.Context) {
	basePath := test.MountAbsolutPath(test.MONGODB_ENVIRONMENT_PATH)

	test.InitializeMongoDBTest()
	pc := test.UseMongoDBContainer(ctx)

	user := &User{}
	datasets := []string{"user1.json", "user2.json"}

	if err := pc.Dataset(user, user.CollectionName(), basePath, datasets...); err != nil {
		logging.Fatal(ctx).Err(err)
	}

	Initialize(ctx)
}
