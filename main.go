package main

import (
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
	"log"
	"os"
)

var Version = "unknown"

func main() {
	app := cli.NewApp()
	app.Name = "drone-plugin-ssh-publisher"
	app.Usage = "Execute the remote SSH command to send the files and artifacts"
	app.Copyright = "Copyright (c) 2021 hb0730"
	app.Authors = []cli.Author{
		{
			Name:  "hb0730",
			Email: "huangbing0730@gmail.com",
		},
	}
	app.Version = Version
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "ssh-key",
			Usage:  "private ssh key",
			EnvVar: "PLUGIN_SSH_KEY,PLUGIN_KEY,SSH_KEY,KEY,INPUT_KEY",
		},
	}
	if _, err := os.Stat("/run/drone/env"); err == nil {
		godotenv.Overload("/run/drone/env")
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
func run(c *cli.Context) error {
	return nil
}
