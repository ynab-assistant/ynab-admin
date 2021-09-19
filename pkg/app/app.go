package app

import (
	"context"
	"github.com/pkg/errors"
	"github.com/ynab-assistant/ynab-admin/pkg/app/api/handlers"
	"github.com/ynab-assistant/ynab-admin/pkg/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"expvar"           // Register the expvar handlers
	_ "net/http/pprof" // Register the pprof handlers
)

func Run(logger *log.Logger, configPath string, build string) error {
	logger.Printf("running the build %s", build)
	expvar.NewString("build").Set(build)

	// ==== Init Configuration
	cfg, err := config.Init(configPath)
	if err != nil {
		return errors.Wrap(err, "cant init configuration for the app")
	}

	// ====  Init Database
	db, err := initDB(cfg)
	if err != nil {
		return errors.Wrap(err, "cant create MongoDB client")
	}

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      handlers.API(build, shutdown, logger, db),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main: API listening on http://127.0.0.1%s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main: %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}

	}

	return nil
}

func initDB(cfg *config.Config) (*mongo.Database, error) {
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
