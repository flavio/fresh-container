package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/flavio/fresh-container/internal/config"
	"github.com/flavio/fresh-container/internal/db"
	"github.com/flavio/fresh-container/internal/workers"

	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

type ApiServer struct {
	backgroundWorker *workers.BackgroundWorker
	db               *db.DB
	cfg              *config.Config
	port             int
	router           *mux.Router
}

func NewApiServer(bw *workers.BackgroundWorker, db *db.DB, port int, cfg *config.Config) (*ApiServer, error) {
	api := ApiServer{
		backgroundWorker: bw,
		cfg:              cfg,
		port:             port,
		router:           mux.NewRouter(),
		db:               db,
	}
	api.initRoutes()

	return &api, nil
}

func (a *ApiServer) ListenAndServe() error {
	loggedRouter := gorilla_handlers.LoggingHandler(os.Stdout, a.router)

	fmt.Printf("Starting server on port %d\n", a.port)

	return http.ListenAndServe(fmt.Sprintf(":%d", a.port), loggedRouter)
}

func (a *ApiServer) initRoutes() {

	a.router.
		Path("/api/v1/check").
		Methods("GET").
		Queries(
			"image", "{image}",
			"constraint", "{constraint}",
		).HandlerFunc(a.Check)

	a.router.
		Path("/api/v1/jobs/{id}").
		Methods("GET").
		HandlerFunc(a.GetJob)

	a.router.
		Path("/api/v1/evaluations/{id}").
		Methods("GET").
		HandlerFunc(a.GetEvaluation)

	a.router.
		Path("/healthz").
		Methods("GET").
		HandlerFunc(a.GetHealthz)
}

func ServeErrorAsJSON(w http.ResponseWriter, statusCode int, err error) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	log.Errorf("Encountered error: %s", err)
	msg := struct {
		Error error
	}{
		err,
	}
	return json.NewEncoder(w).Encode(msg)
}
