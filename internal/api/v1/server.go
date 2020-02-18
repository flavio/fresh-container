package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/flavio/stale-container/internal/config"
	"github.com/flavio/stale-container/internal/db"
	"github.com/flavio/stale-container/internal/workers"

	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type ApiServer struct {
	backgrondWorker *workers.BackgroundWorker
	db              *db.DB
	cfg             *config.Config
	port            int
	router          *mux.Router
}

func NewApiServer(bw *workers.BackgroundWorker, db *db.DB, port int, cfg *config.Config) (*ApiServer, error) {
	api := ApiServer{
		backgrondWorker: bw,
		cfg:             cfg,
		port:            port,
		router:          mux.NewRouter(),
		db:              db,
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
}

func ServeErrorAsJSON(w http.ResponseWriter, statusCode int, err error) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	msg := struct {
		Error error
	}{
		err,
	}
	return json.NewEncoder(w).Encode(msg)
}
