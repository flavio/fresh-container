package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blang/semver"
	"github.com/flavio/fresh-container/pkg/fresh_container"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (a *ApiServer) Check(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// tagPrefix is optional so have to retrieve it explicitely
	tagPrefix := r.URL.Query().Get("tagPrefix")

	log.WithFields(log.Fields{
		"image":      vars["image"],
		"constraint": vars["constraint"],
		"tagPrefix":  tagPrefix,
		"host":       r.Host,
	}).Debug("GET check")

	image, err := fresh_container.NewImage(vars["image"], tagPrefix)
	if err != nil {
		ServeErrorAsJSON(w, http.StatusBadRequest, err)
		return
	}

	_, err = semver.ParseRange(vars["constraint"])
	if err != nil {
		ServeErrorAsJSON(w, http.StatusBadRequest, err)
		return
	}

	tags, err := a.db.GetImageTags(image)
	if err != nil {
		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	if len(tags) == 0 {
		// No tags - queue the job
		id, err := a.backgroundWorker.AddJob(vars["image"], vars["constraint"], tagPrefix)
		if err != nil {
			ServeErrorAsJSON(w, http.StatusInternalServerError, err)
			return
		}

		serveJobAcceptedResponse(id, w)
		return
	}

	// prefixes have already been stripped since we got this from cache
	if err = image.SetTagVersions(tags, true, true); err != nil {
		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	evaluation, err := image.EvalUpgrade(vars["constraint"])

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(evaluation)
}

func serveJobAcceptedResponse(jobId string, w http.ResponseWriter) {
	w.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", jobId))
	w.WriteHeader(http.StatusAccepted)
}
