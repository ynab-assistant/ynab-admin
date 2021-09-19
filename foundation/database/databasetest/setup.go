package databasetest

import (
	"context"
	"fmt"
	"github.com/ynab-assistant/ynab-admin/foundation/database"
	"github.com/ynab-assistant/ynab-admin/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"testing"
	"time"
)

func Setup(t *testing.T) (*mongo.Database, func()) {
	t.Helper()

	c := startContainer(t)
	mongoURI := fmt.Sprintf("mongodb://%s", c.Host)

	db, err := database.InitDB(&config.Config{
		Mongo: config.MongoConfig{
			Name:     "test",
			URI:      mongoURI,
			User:     "root",
			Password: "root",
		},
	})

	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("waiting for database to be ready")

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Client().Ping(ctx, readpref.Primary())
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		stopContainer(t, c)
		t.Fatalf("waiting for ping from database: %v", pingError)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		t.Helper()
		db.Client().Disconnect(ctx)
		stopContainer(t, c)
	}

	return db, teardown
}
