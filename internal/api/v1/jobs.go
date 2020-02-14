package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flavio/stale-container/internal/db"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type jobStatus struct {
	Status string `json:"status"`
}

func (a *ApiServer) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.WithFields(log.Fields{
		"id":   vars["id"],
		"host": r.Host,
	}).Debug("GET job")

	_, err := a.db.GetEvaluation(vars["id"])
	if err != nil {
		if err == db.ErrorEvaluationNotFound {
			// maybe the job hasn't been picked up by the worker yet
			queued, err := a.db.IsJobQueued(vars["id"])
			if err != nil {
				ServeErrorAsJSON(w, http.StatusInternalServerError, err)
				return
			}

			if queued {
				res := jobStatus{Status: "pending"}

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)

				json.NewEncoder(w).Encode(res)
				return
			} else {
				// job has gone lost
				err = fmt.Errorf("Job not found")
				ServeErrorAsJSON(w, http.StatusNotFound, err)
				return
			}
		}

		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/evaluations/%s", vars["id"]))
	w.WriteHeader(http.StatusSeeOther)
}
