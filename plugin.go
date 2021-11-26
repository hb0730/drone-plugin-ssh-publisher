package main

import (
	"errors"
	"fmt"
	"github.com/appleboy/easyssh-proxy"
	"github.com/mholt/archiver/v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errMissingHost          = errors.New("Error: missing server host")
	errMissingPasswordOrKey = errors.New("Error: can't connect without a private SSH key or password")
	errSetPasswordandKey    = errors.New("can't set password and key at the same time")
)

// Plugin structure
type Plugin struct {
	sshConfig *easyssh.MakeConfig
	Writer    io.Writer
	Host      Host
	Transfer  SshTransfer
	Command   Commands
}

type Commands struct {
	Commands       []string
	CommandTimeout time.Duration
	Command        []string
}

type fileList struct {
	Ignore []string
	Source []string
}

func trimPath(keys []string) []string {
	var newKeys []string

	for _, value := range keys {
		value = strings.Trim(value, " ")
		if len(value) == 0 {
			continue
		}

		newKeys = append(newKeys, value)
	}

	return newKeys
}

func globList(paths []string) fileList {
	var list fileList
	for _, pattern := range paths {

		ignore := false
		pattern = strings.Trim(pattern, " ")
		if string(pattern[0]) == "!" {
			pattern = pattern[1:]
			ignore = true
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Glob error for %q: %s\n", pattern, err)
			continue
		}
		if ignore {
			list.Ignore = append(list.Ignore, matches...)
		} else {
			list.Source = append(list.Source, matches...)
		}
	}
	return list
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

type SshTransfer struct {
	Target       []string
	Source       []string
	RemovePrefix []string
	CleanRemote  bool

	destFile string
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
	err := p.execTransfer()
	if err != nil {
		return err
	}
	return p.execCommands()
}
func (p *Plugin) log(message ...interface{}) {
	if p.Writer == nil {
		p.Writer = os.Stdout
	}
	fmt.Fprintf(p.Writer, "%s: %s", p.Host.Host, fmt.Sprintln(message...))
}
func (p *Plugin) execCommands() error {

	return nil
}

func (p *Plugin) execTransfer() error {
	if len(p.Transfer.Source) == 0 || len(p.Transfer.Target) == 0 {
		return nil
	}
	// 获取所有的文件及文件夹
	files := globList(trimPath(p.Transfer.Source))
	// 临时文件
	p.Transfer.destFile = fmt.Sprintf("%d.tar", time.Now().Unix())
	//文件临时文件夹
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	tar := filepath.Join(dir, p.Transfer.destFile)
	p.log("tar all files into " + tar)
	// 压缩文件及文件夹
	err = archiver.Archive(files.Source, tar)
	if err != nil {
		return err
	}
	ssh := p.sshConfig
	// 传输
	err = ssh.Scp(tar, p.Transfer.destFile)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(p.Transfer.Target))
	errChannel := make(chan error)
	finished := make(chan struct{})
	for _, target := range p.Transfer.Target {
		go p.operationTargetDir(target, errChannel, &wg)
	}
	go func() {
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case err := <-errChannel:
		if err != nil {
			p.log("drone-plugin-ssh-publisher error: ", err)
			p.log("drone-plugin-ssh-publisher error: remove all target tmp file")
			if err := p.removeDestFile(); err != nil {
				return err
			}
			return err
		}
	}
	// 最后删除临时文件
	return p.removeDestFile()
}

func (p *Plugin) operationTargetDir(target string, errChannel chan error, wg *sync.WaitGroup) {
	var cmd string
	// 是否清理远程目录
	if p.Transfer.CleanRemote {
		cmd = fmt.Sprintf("rm -rf %s", target)
		p.log("Remove target folder:", target, " ,command: ", cmd)
		_, _, _, err := p.sshConfig.Run(cmd, p.Command.CommandTimeout)
		if err != nil {
			errChannel <- err
			return
		}
	}
	// mkdir path
	cmd = fmt.Sprintf("mkdir -p %s", target)
	p.log("create folder", target, " command:", cmd)
	_, errStr, _, err := p.sshConfig.Run(cmd, p.Command.CommandTimeout)
	if err != nil {
		errChannel <- err
	}
	if len(errStr) != 0 {
		errChannel <- fmt.Errorf(errStr)
		return
	}
	// untar file
	cmd = fmt.Sprintf("tar -xf %s -C %s", p.Transfer.destFile, target)
	p.log("untar file", p.Transfer.destFile, " command:", cmd)
	outStr, errStr, _, err := p.sshConfig.Run(cmd, p.Command.CommandTimeout)
	if outStr != "" {
		p.log("output: ", outStr)
	}
	if errStr != "" {
		p.log("error: ", errStr)
	}
	if err != nil {
		errChannel <- err
		return
	}
	for _, removePrefix := range p.Transfer.RemovePrefix {
		if len(removePrefix) == 0 {
			break
		}
		cmd = fmt.Sprintf("cd %s \n mv -f %s/* . \n rm -rf %s", target, removePrefix, removePrefix)
		p.log("remove prefix ", removePrefix, " command: ", cmd)
		outStr, errStr, _, err = p.sshConfig.Run(cmd, p.Command.CommandTimeout)
		if outStr != "" {
			p.log("remove prefix output: ", outStr)
		}
		if errStr != "" {
			p.log("remove prefix error: ", errStr)
		}
		if err != nil {
			errChannel <- err
			return
		}
	}
	wg.Done()

}
func (p *Plugin) removeDestFile() error {
	cmd := fmt.Sprintf("rm -rf %s", p.Transfer.destFile)
	p.log("remove file", p.Transfer.destFile, " command: ", cmd)
	_, errStr, _, err := p.sshConfig.Run(cmd, p.Command.CommandTimeout)
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return fmt.Errorf(errStr)
	}
	return nil
}
