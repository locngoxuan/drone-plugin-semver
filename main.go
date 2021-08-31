package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
		cli.BoolFlag{
			Name:   "require-action",
			Usage:  "return error if action is empty", //release, patch
			EnvVar: "PLUGIN_REQUIRE_ACTION",
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
			Action:           strings.TrimSpace(c.String("action")),
			Output:           c.StringSlice("output"),
			DroneBuildNumber: os.Getenv("DRONE_BUILD_NUMBER"),
			RequireAction:    c.Bool("require-action"),
		},
	}
	if strings.TrimSpace(plugin.Config.Action) == "" {
		if plugin.Config.RequireAction {
			return fmt.Errorf(`action must not empty`)
		}
	}
	fmt.Println("action: ", plugin.Config.Action)
	fmt.Println("build number: ", plugin.Config.DroneBuildNumber)
	fmt.Println("output: ", plugin.Config.Output)
	fmt.Println("require-action: ", plugin.Config.RequireAction)
	return plugin.Exec()
}
