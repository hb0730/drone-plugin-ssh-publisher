package main

import (
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
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
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "PLUGIN_DEBUG,DEBUG,INPUT_DEBUG",
		},
		cli.StringFlag{
			Name:   "ssh-key",
			Usage:  "private ssh key",
			EnvVar: "PLUGIN_SSH_KEY,PLUGIN_KEY,SSH_KEY,KEY,INPUT_KEY",
		},
		cli.StringFlag{
			Name:   "ssh-passphrase",
			Usage:  "The purpose of the passphrase is usually to encrypt the private key.",
			EnvVar: "PLUGIN_SSH_PASSPHRASE,PLUGIN_PASSPHRASE,SSH_PASSPHRASE,PASSPHRASE,INPUT_PASSPHRASE",
		},
		cli.StringFlag{
			Name:   "key-path,i",
			Usage:  "ssh private key path",
			EnvVar: "PLUGIN_KEY_PATH,SSH_KEY_PATH,INPUT_KEY_PATH",
		},
		cli.StringFlag{
			Name:   "username,user,u",
			Usage:  "connect as user",
			EnvVar: "PLUGIN_USERNAME,PLUGIN_USER,SSH_USERNAME,USERNAME,INPUT_USERNAME",
			Value:  "root",
		},
		cli.StringFlag{
			Name:   "password,P",
			Usage:  "user password",
			EnvVar: "PLUGIN_PASSWORD,SSH_PASSWORD,PASSWORD,INPUT_PASSWORD",
		},
		cli.StringSliceFlag{
			Name:     "host,H",
			Usage:    "connect to host",
			EnvVar:   "PLUGIN_HOST,SSH_HOST,HOST,INPUT_HOST",
			FilePath: ".host",
		},
		cli.IntFlag{
			Name:   "port,p",
			Usage:  "connect to port",
			EnvVar: "PLUGIN_PORT,SSH_PORT,PORT,INPUT_PORT",
			Value:  22,
		},
		cli.DurationFlag{
			Name:   "timeout,t",
			Usage:  "connection timeout",
			EnvVar: "PLUGIN_TIMEOUT,SSH_TIMEOUT,TIMEOUT,INPUT_TIMEOUT",
			Value:  30 * time.Second,
		},
		cli.DurationFlag{
			Name:   "command.timeout,T",
			Usage:  "command timeout",
			EnvVar: "PLUGIN_COMMAND_TIMEOUT,SSH_COMMAND_TIMEOUT,COMMAND_TIMEOUT,INPUT_COMMAND_TIMEOUT",
			Value:  10 * time.Minute,
		},
		cli.StringSliceFlag{
			Name:   "script,s",
			Usage:  "execute commands",
			EnvVar: "PLUGIN_SCRIPT,SSH_SCRIPT,SCRIPT",
		},
		cli.StringSliceFlag{
			Name:   "target, t",
			Usage:  "Target path on the server",
			EnvVar: "PLUGIN_TARGET,SCP_TARGET,TARGET,INPUT_TARGET",
		},
		cli.StringSliceFlag{
			Name:   "source, s",
			Usage:  "scp file list",
			EnvVar: "PLUGIN_SOURCE,SCP_SOURCE,SOURCE,INPUT_SOURCE",
		},
		cli.StringSliceFlag{
			Name:   "removePrefix, rp",
			Usage:  "remove target path prefix",
			EnvVar: "PLUGIN_REMOVE_PREFIX,SCP_REMOVE_PREFIX,REMOVE_PREFIX,INPUT_REMOVE_PREFIX",
		},
		cli.BoolFlag{
			Name:   "rm, r",
			Usage:  "remove target folder before upload data",
			EnvVar: "PLUGIN_RM,SCP_RM,RM,INPUT_RM",
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
	p := &Plugin{
		Debug: c.Bool("debug"),
		Host: Host{
			Host:       c.String("host"),
			Username:   c.String("username"),
			Password:   c.String("password"),
			Port:       c.Int("port"),
			Key:        c.String("ssh-key"),
			KeyPath:    c.String("key-path"),
			Passphrase: c.String("ssh-passphrase"),
			Timeout:    c.Duration("timeout"),
		},
		Artifact: Artifacts{
			Source:       c.StringSlice("source"),
			Target:       c.StringSlice("target"),
			RemovePrefix: c.StringSlice("removePrefix"),
			CleanRemote:  c.Bool("rm"),
		},
		Command: Command{
			Commands:       c.StringSlice("script"),
			CommandTimeout: c.Duration("command.timeout"),
		},
		Writer: os.Stdout,
	}
	return p.Exec()
}
