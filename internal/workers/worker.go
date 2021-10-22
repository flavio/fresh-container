package workers

import (
	"context"

	"github.com/flavio/fresh-container/internal/config"
	"github.com/flavio/fresh-container/internal/db"

	"github.com/google/uuid"
	"github.com/vmihailenco/taskq/v2"
	"github.com/vmihailenco/taskq/v2/memqueue"
)

type BackgroundWorker struct {
	config       *config.Config
	db           *db.DB
	ctx          context.Context
	queueFactory taskq.Factory
	queue        taskq.Queue
	task         *taskq.Task
}

func (w *BackgroundWorker) AddJob(image, constraint, tagPrefix string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	if err = w.queue.Add(w.task.WithArgs(w.ctx, id.String(), image, constraint, tagPrefix)); err != nil {
		return "", err
	}
	if err := w.db.SetJobQueued(id.String()); err != nil {
		return "", err
	}

	return id.String(), nil
}

func NewBackgroungWorker(config *config.Config, db *db.DB) *BackgroundWorker {
	bw := &BackgroundWorker{
		db:     db,
		config: config,
	}

	bw.queueFactory = memqueue.NewFactory()
	bw.queue = bw.queueFactory.RegisterQueue(
		&taskq.QueueOptions{
			Name: "constraint-solver",
		},
	)

	bw.task = taskq.RegisterTask(&taskq.TaskOptions{
		Name:    "fetch-tags-and-solve-constraint",
		Handler: bw.ProcessJob,
	})

	return bw
}
