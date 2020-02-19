package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/flavio/fresh-container/internal/api/v1"
	"github.com/flavio/fresh-container/internal/config"
	"github.com/flavio/fresh-container/internal/db"
	"github.com/flavio/fresh-container/internal/workers"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func RunServer(c *cli.Context) error {
	cfg := config.NewConfig()
	var err error

	// ensure cleanup is done when the program is terminated
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-sigChannel
		os.Exit(0)
	}()

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	if c.String("config") != "" {
		cfg, err = config.NewFromFile(c.String("config"))
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	db, err := db.NewDB(&cfg)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	go dbGarbageCollectorLoop(db)

	// BackgroundWorker used to handle jobs
	bw := workers.NewBackgroungWorker(&cfg, db)

	server, err := api.NewApiServer(bw, db, c.Int("port"), &cfg)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	err = server.ListenAndServe()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

func dbGarbageCollectorLoop(db *db.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
	again:
		err := db.GC()
		if err == nil {
			goto again
		}
	}
}
