package workers

import (
	"github.com/flavio/fresh-container/pkg/fresh_container"

	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (w *BackgroundWorker) _handleError(ctx context.Context, tagPrefix string, id string, img string, constraint string, action string, err error) {
	log.WithFields(log.Fields{
		"action":     action,
		"id":         id,
		"image":      img,
		"constraint": constraint,
		"tagPrefix":  tagPrefix,
		"error":      err,
	}).Error("worker.ProcessJob")
	msg := struct {
		Error string
	}{
		fmt.Sprint(err),
	}
	encodedResult, err1 := json.Marshal(msg)
	if err1 != nil {
		log.WithFields(log.Fields{
			"action":     "_handleError:Marshal",
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"tagPrefix":  tagPrefix,
			"error":      err,
		}).Error("worker.ProcessJob")
		return
	}
	if err2 := w.db.SetEvaluation(id, encodedResult, fmt.Sprint(http.StatusInternalServerError)); err2 != nil {
		log.WithFields(log.Fields{
			"action":     "_handleError:SetEvaluation",
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"tagPrefix":  tagPrefix,
			"error":      err,
		}).Error("worker.ProcessJob")
	}

}

func (w *BackgroundWorker) ProcessJob(ctx context.Context, id, img, constraint, tagPrefix string) error {
	image, err := fresh_container.NewImage(img, tagPrefix)
	if err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "NewImage", err)
		return err
	}

	// reach to external registry to fetch tags
	if err = image.FetchTags(ctx, w.config); err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "FetchTags", err)
		return err
	}

	// save tags into DB
	tagsString := []string{}
	for _, tag := range image.TagVersions {
		tagsString = append(tagsString, tag.String())
	}
	if err = w.db.SetImageTags(image, tagsString); err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "save_tags", err)
		return err
	} else {
		log.WithFields(log.Fields{
			"action":     "save_tags",
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"tagPrefix":  tagPrefix,
			"tags":       tagsString,
		}).Debug("worker.ProcessJob")
	}

	evaluation, err := image.EvalUpgrade(constraint)
	if err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "EvalUpgrade", err)
		return err
	}

	encodedResult, err := json.Marshal(evaluation)
	if err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "Marshal", err)
		return err
	}

	if err = w.db.SetEvaluation(id, encodedResult, ""); err != nil {
		w._handleError(ctx, tagPrefix, id, img, constraint, "SetEvaluation", err)
		return err
	}

	return err
}
