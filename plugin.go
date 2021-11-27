package main

import (
	"errors"
	"fmt"
	"github.com/appleboy/easyssh-proxy"
	"io"
	"os"
	"strconv"
	"time"
)

var (
	errMissingHost          = errors.New("Error: missing server host ")
	errMissingPasswordOrKey = errors.New("Error: can't connect without a private SSH key or password ")
	errSetPasswordandKey    = errors.New("can't set password and key at the same time ")
	errCommandTimeOut       = errors.New("Error: command timeout ")
)

// Plugin structure
type Plugin struct {
	Debug     bool
	sshConfig *easyssh.MakeConfig
	Writer    io.Writer
	Host      Host
	Artifact  Artifacts
	Command   Commands
}

type Commands struct {
	Commands       []string
	CommandTimeout time.Duration
}

// Host  server config
type Host struct {
	Key        string
	Passphrase string
	KeyPath    string
	Username   string
	Password   string
	Host       string
	Port       int
	Timeout    time.Duration
}

type Artifacts struct {
	Target       []string
	Source       []string
	RemovePrefix []string
	CleanRemote  bool

	tempFile string
}

func (p *Plugin) Exec() error {
	if len(p.Host.Host) == 0 {
		return errMissingHost
	}
	if len(p.Host.Key) == 0 && len(p.Host.Password) == 0 && len(p.Host.KeyPath) == 0 {
		return errMissingPasswordOrKey
	}
	if len(p.Host.Key) != 0 && len(p.Host.Password) != 0 {
		return errSetPasswordandKey
	}
	p.sshConfig = &easyssh.MakeConfig{
		User:       p.Host.Username,
		Server:     p.Host.Host,
		Password:   p.Host.Password,
		Port:       strconv.Itoa(p.Host.Port),
		Key:        p.Host.Key,
		KeyPath:    p.Host.KeyPath,
		Passphrase: p.Host.Passphrase,
		Timeout:    p.Host.Timeout,
	}
	err := p.execArtifact()
	if err != nil {
		return err
	}
	return p.execCommand()
}
func (p *Plugin) log(message ...interface{}) {
	if p.Writer == nil {
		p.Writer = os.Stdout
	}
	fmt.Fprintf(p.Writer, "%s", fmt.Sprintln(message...))
}
