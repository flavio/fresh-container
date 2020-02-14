package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flavio/stale-container/internal/config"
	"github.com/flavio/stale-container/pkg/stale_container"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	ValidOututFormats = []string{"text", "json"}
)

func isOutputFormatValid(format string) bool {
	for _, f := range ValidOututFormats {
		if f == format {
			return true
		}
	}

	return false
}

func CheckImage(c *cli.Context) error {
	var err error
	var evaluation stale_container.ImageUpgradeEvaluationResponse
	constraint := c.String("constraint")

	if c.NArg() != 1 {
		return cli.NewExitError("Wrong usage", 1)
	}

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	output := c.String("output")
	if !isOutputFormatValid(output) {
		err := fmt.Errorf(
			"Invalid output format: %s. Valid ones are %+v",
			output,
			ValidOututFormats)
		return cli.NewExitError(err, 1)
	}

	if c.String("server") == "" {
		evaluation, err = localEvaluation(
			c.Args().Get(0),
			constraint,
			c.String("config"),
			c.Context)
	} else {
		if c.String("config") != "" {
			log.Warn("`config` flag is ignored when the `server` is used at the same time")
		}
	}
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	switch output {
	case "text":
		if !evaluation.Stale {
			fmt.Printf(
				"%s is already the latest version available that satisfies the %s constraint\n",
				evaluation.Image, evaluation.Constraint)
		} else {
			err := fmt.Errorf(
				"The '%s' container image can be upgraded from the '%s' tag to the '%s' one and still satisfy the '%s' constraint.",
				evaluation.Image,
				evaluation.CurrentVersion,
				evaluation.NextVersion,
				evaluation.Constraint)
			return cli.NewExitError(err, 1)
		}
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(evaluation); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	return nil
}

func localEvaluation(image, constraint, configFile string, ctx context.Context) (evaluation stale_container.ImageUpgradeEvaluationResponse, err error) {
	cfg := config.NewConfig()
	if configFile != "" {
		cfg, err = config.NewFromFile(configFile)
		if err != nil {
			return
		}
	}

	img, err := stale_container.NewImage(image)
	if err != nil {
		return
	}

	err = img.FetchTags(ctx, &cfg)
	if err != nil {
		return
	}

	return img.EvalUpgrade(constraint)
}
