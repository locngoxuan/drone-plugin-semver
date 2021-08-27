package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var (
	version = "unknown"
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "semver plugin"
	app.Usage = "semver plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "action",
			Usage:  "specify versioning action", //release, patch
			EnvVar: "PLUGIN_ACTION",
		},
		cli.StringSliceFlag{
			Name:   "output",
			Usage:  "specify output of semver",
			Value:  &cli.StringSlice{".drone.semver"},
			EnvVar: "PLUGIN_OUTPUT",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := &Plugin{
		Config: Config{
			Src:              "VERSION",
			Action:           c.String("action"),
			Output:           c.StringSlice("output"),
			DroneBuildNumber: os.Getenv("DRONE_BUILD_NUMBER"),
			DroneBuildRef:    os.Getenv("DRONE_COMMIT_REF"),
			DroneBuildBranch: os.Getenv("DRONE_COMMIT_BRANCH"),
		},
	}
	return plugin.Exec()
}
