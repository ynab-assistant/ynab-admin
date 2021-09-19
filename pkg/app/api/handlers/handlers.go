package handlers

import (
	"github.com/ynab-assistant/ynab-admin/foundation/web"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, logger *log.Logger, db *mongo.Database) http.Handler {
	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown)

	// Register debug check endpoints.
	cg := checkGroup{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/v1/readiness", cg.readiness)
	app.Handle(http.MethodGet, "/v1/liveness", cg.liveness)

	return app
}
