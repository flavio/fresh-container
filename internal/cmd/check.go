package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flavio/stale-container/internal/common"
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
	var cfg config.Config
	var err error
	constraint := c.String("constraint")

	if c.NArg() != 1 {
		return cli.NewExitError("Wrong usage", 1)
	}

	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.String("config") != "" {
		cfg, err = config.NewFromFile(c.String("config"))
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	output := c.String("output")
	if !isOutputFormatValid(output) {
		err := fmt.Errorf(
			"Invalid output format: %s. Valid ones are %+v",
			output,
			ValidOututFormats)
		return cli.NewExitError(err, 1)
	}

	image, err := stale_container.NewImage(c.Args().Get(0))
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	err = image.FetchTags(c.Context, &cfg)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	nextVer, err := image.EvalUpgrade(constraint)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	switch output {
	case "text":
		if nextVer.LTE(image.TagVersion) {
			fmt.Printf(
				"%s is already the latest version available that satisfies the %s constraint\n",
				image, constraint)
		} else {
			fmt.Printf("%s can be updated to the %s release\n", image, nextVer.String())
			return cli.NewExitError(fmt.Errorf("Image outdated"), 1)
		}
	case "json":
		res := types.NewCheckResponse(image, constraint, nextVer)

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(res); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	return nil
}
