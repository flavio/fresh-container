package workers

import (
	"github.com/flavio/fresh-container/pkg/fresh_container"

	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func (w *BackgroundWorker) ProcessJob(ctx context.Context, id, img, constraint string) error {
	image, err := fresh_container.NewImage(img)
	if err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	}

	// reach to external registry to fetch tags
	if err = image.FetchTags(ctx, w.config); err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	}

	// save tags into DB
	tagsString := []string{}
	for _, tag := range image.TagVersions {
		tagsString = append(tagsString, tag.String())
	}
	if err = w.db.SetImageTags(image, tagsString); err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	} else {
		log.WithFields(log.Fields{
			"action":     "save_tags",
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"tags":       tagsString,
		}).Debug("worker.ProcessJob")
	}

	evaluation, err := image.EvalUpgrade(constraint)
	if err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	}

	encodedResult, err := json.Marshal(evaluation)
	if err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	}

	if err = w.db.SetEvaluation(id, encodedResult); err != nil {
		log.WithFields(log.Fields{
			"id":         id,
			"image":      img,
			"constraint": constraint,
			"error":      err,
		}).Error("worker.ProcessJob")
		return err
	}

	return err
}
