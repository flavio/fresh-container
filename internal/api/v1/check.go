package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blang/semver"
	"github.com/flavio/stale-container/internal/common"
	"github.com/flavio/stale-container/pkg/stale_container"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (a *ApiServer) Check(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.WithFields(log.Fields{
		"image":      vars["image"],
		"constraint": vars["constraint"],
		"host":       r.Host,
	}).Debug("GET check")

	image, err := stale_container.NewImage(vars["image"])
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
		id, err := a.backgrondWorker.AddJob(vars["image"], vars["constraint"])
		if err != nil {
			ServeErrorAsJSON(w, http.StatusInternalServerError, err)
			return
		}

		serveJobAcceptedResponse(id, w)
		return
	}

	if err = image.SetTagVersions(tags, true); err != nil {
		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	nextTagVersion, err := image.EvalUpgrade(vars["constraint"])
	if err = image.SetTagVersions(tags, true); err != nil {
		ServeErrorAsJSON(w, http.StatusInternalServerError, err)
		return
	}

	res := types.NewCheckResponse(image, vars["constraint"], nextTagVersion)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func serveJobAcceptedResponse(jobId string, w http.ResponseWriter) {
	w.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", jobId))
	w.WriteHeader(http.StatusAccepted)
}
