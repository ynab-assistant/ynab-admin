package database

import (
	"context"
	"github.com/ynab-assistant/ynab-admin/pkg/config"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// InitDB creates a new MongoDB database for testing purposes
func InitDB(cfg *config.Config) (*mongo.Database, error) {
	mongoClient, err := newMongoClient(cfg.Mongo)
	if err != nil {
		return nil, err
	}

	return mongoClient.Database(cfg.Mongo.Name), nil
}

func newMongoClient(cfg config.MongoConfig) (*mongo.Client, error) {

	credential := options.Credential{
		Username: cfg.User,
		Password: cfg.Password,
	}
	clientOpts := options.Client().ApplyURI(cfg.URI).SetAuth(credential)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)

	if err != nil {
		return nil, errors.Wrapf(err, "error whicle connecting to MongoDB.\n\tURI:\t%s \n\tUser: %s", cfg.URI, cfg.User)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, errors.Wrapf(err, " can't ping MongoDB.\n\tURI:\t%s \n\tUser: %s", cfg.URI, cfg.User)
	}

	return client, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *mongo.Database) error {

	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	// db.runCommand( { serverStatus: 1 } ).ok
	var dbStatus struct {
		Ok int `bson:"ok"`
	}

	if err := db.RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&dbStatus); err != nil {
		return err
	}

	if dbStatus.Ok != 1 {
		return errors.New("mongoDB is not ready")
	}
	return nil
}
