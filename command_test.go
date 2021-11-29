package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPlugin_execCommand(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv("port"))
	commands := strings.Split(os.Getenv("cmds"), ",")
	//cleanRemote, _ := strconv.ParseBool(os.Getenv("cleanRemote"))
	plugin := &Plugin{
		Debug: true,
		Host: Host{
			Username: os.Getenv("username"),
			Password: os.Getenv("password"),
			Host:     os.Getenv("host"),
			Port:     port,
		},
		Command: Command{
			CommandTimeout: 10 * time.Minute,
			Commands:       commands,
		},
	}
	err := plugin.Exec()
	if err != nil {
		t.Errorf("error:%s", err.Error())
	}
}
