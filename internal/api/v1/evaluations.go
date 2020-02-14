package api

import (
	"net/http"

	"github.com/flavio/stale-container/internal/db"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (a *ApiServer) GetEvaluation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.WithFields(log.Fields{
		"id":   vars["id"],
		"host": r.Host,
	}).Debug("GET evaluation")

	res, err := a.db.GetEvaluation(vars["id"])
	if err != nil {
		if err == db.ErrorEvaluationNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}
