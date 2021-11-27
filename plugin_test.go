package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_globList(t *testing.T) {
	paths := strings.Split(os.Getenv("source"), ",")
	files := globList(paths)
	t.Logf("%v", files)
}
func TestTempDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("crate tmp dir error:%s", err.Error())
	}
	t.Logf("tmp dir:%s", tempDir)
}

func TestPlugin_sshTransfer(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv("port"))
	cleanRemote, _ := strconv.ParseBool(os.Getenv("cleanRemote"))
	plugin := &Plugin{
		Debug: true,
		Host: Host{
			Username: os.Getenv("username"),
			Password: os.Getenv("password"),
			Host:     os.Getenv("host"),
			Port:     port,
		},
		Artifact: Artifacts{
			Source:       strings.Split(os.Getenv("source"), ","),
			Target:       strings.Split(os.Getenv("target"), ","),
			RemovePrefix: strings.Split(os.Getenv("removePrefix"), ","),
			CleanRemote:  cleanRemote,
		},
		Command: Command{
			CommandTimeout: 10 * time.Minute,
		},
	}
	err := plugin.Exec()
	if err != nil {
		t.Errorf("error:%s", err.Error())
	}
}

func TestPlugin_execCommands(t *testing.T) {
	port, _ := strconv.Atoi(os.Getenv("port"))
	commands := strings.Split(os.Getenv("cmds"), ",")
	cleanRemote, _ := strconv.ParseBool(os.Getenv("cleanRemote"))
	plugin := &Plugin{
		Debug: true,
		Host: Host{
			Username: os.Getenv("username"),
			Password: os.Getenv("password"),
			Host:     os.Getenv("host"),
			Port:     port,
		},
		Artifact: Artifacts{
			Source:       strings.Split(os.Getenv("source"), ","),
			Target:       strings.Split(os.Getenv("target"), ","),
			RemovePrefix: strings.Split(os.Getenv("removePrefix"), ","),
			CleanRemote:  cleanRemote,
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
