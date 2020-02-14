package main

import (
	"os"

	"github.com/flavio/stale-container/internal/cmd"

	"github.com/urfave/cli/v2"
)

const VERSION = "0.1.0"

func main() {
	app := &cli.App{
		Name:    "stale-container",
		Usage:   "Tool to find stale containers",
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "Configuration file",
				EnvVars: []string{"STALE_CONFIG_FILE"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable extra debugging",
				EnvVars: []string{"STALE_DEBUG"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "check",
				Usage: "Check if the specified image is stale",
				Description: `Given a user defined expiration rule checks if the specified container is stale.

The image tags - both the current one and the remote ones - must respect semantic versioning (https://semver.org/).

The expiration constraint must be formulated to be understood by https://github.com/blang/semver#ranges

A condition is composed of an operator and a version. The supported operators are:

  * '<1.0.0' Less than 1.0.0
  * '<=1.0.0' Less than or equal to 1.0.0
  * '>1.0.0' Greater than 1.0.0
  * '>=1.0.0' Greater than or equal to 1.0.0
  * '1.0.0', '=1.0.0', '==1.0.0' Equal to 1.0.0
  * '!1.0.0', '!=1.0.0' Not equal to 1.0.0. Excludes version 1.0.0.

Note that spaces between the operator and the version will be gracefully tolerated.

Ranges can be linked by logical AND:

  * '>1.0.0 <2.0.0' would match between both ranges, so 1.1.1 and 1.8.7 but not 1.0.0 or 2.0.0
  * '>1.0.0 <3.0.0 !2.0.3-beta.2' would match every version between 1.0.0 and 3.0.0 except 2.0.3-beta.2

Ranges can also be linked by logical OR:

  *  '<2.0.0 || >=3.0.0' would match 1.x.x and 3.x.x but not 2.x.x

AND has a higher precedence than OR. It's not possible to use brackets.

Ranges can be combined by both AND and OR

  *  '>1.0.0 <2.0.0 || >3.0.0 !4.2.1' would match 1.2.3, 1.9.9, 3.1.1, but not 4.2.1, 2.1.1

Example:

$ stale-container check --constraint ">= 1.5.0 < 1.6.0" "influxdb:1.5.0"
`,
				UsageText: "stale-container check --constraint <STALE_CONSTRAINT> <IMAGE>",
				Action:    cmd.CheckImage,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "constraint",
						Usage:    "Expiration constraint - must follow semver rules",
						EnvVars:  []string{"STALE_CONSTRAINT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (json,text)",
						EnvVars: []string{"STALE_CONSTRAINT"},
						Value:   "text",
					},
				},
			},
			{
				Name:        "server",
				Usage:       "Run a simple REST API",
				Description: "Run simple web server that can be used to find stale containers",
				UsageText:   "stale-container server --port <STALE_PORT>",
				Action:      cmd.RunServer,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   5000,
						Usage:   "Listen to port",
						EnvVars: []string{"STALE_PORT"},
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
