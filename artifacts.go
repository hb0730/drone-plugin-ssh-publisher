package main

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

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
func (p *Plugin) execArtifact() error {
	if len(p.Artifact.Source) == 0 || len(p.Artifact.Target) == 0 {
		if p.Debug {
			p.log("[artifact] source or target file is empty")
		}
		return nil
	}
	files := globList(trimPath(p.Artifact.Source))
	p.Artifact.tempFile = fmt.Sprintf("%d.tar", time.Now().Unix())
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	tempTar := filepath.Join(tempDir, p.Artifact.tempFile)
	p.log("[artifact] tar all files to ", tempTar)
	err = archiver.Archive(files.Source, tempTar)
	if err != nil {
		return err
	}
	p.log("[artifact] transfer ....")
	err = p.sshConfig.Scp(tempTar, p.Artifact.tempFile)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(p.Artifact.Target))
	errChannel := make(chan error)
	finished := make(chan struct{})
	for _, target := range p.Artifact.Target {
		go p.transferArtifact(target, errChannel, &wg)
	}
	go func() {
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case err := <-errChannel:
		if err != nil {
			p.log("[artifact] transfer error ", err)
			if err := p.removeTempFile(); err != nil {
				p.log("")
				return err
			}
			return err
		}
	}
	if err := p.removeTempFile(); err != nil {
		return err
	}
	p.log("=============")
	p.log("✅ Successfully executed transfer data to host")
	p.log("=============")
	return nil
}
func (p *Plugin) transferArtifact(target string, errChannel chan error, wg *sync.WaitGroup) {
	if p.Artifact.CleanRemote {
		p.log("[artifact] clean remote dir ", target)
		_, errStr, _, err := p.sshConfig.Run(fmt.Sprintf("rm -rf %s", target), p.Command.CommandTimeout)
		if err != nil {
			errChannel <- err
			return
		}
		if len(errStr) != 0 {
			errChannel <- fmt.Errorf(errStr)
			return
		}
	}
	// mkdir
	p.log("[artifact] create dir ", target)
	_, errStr, _, err := p.sshConfig.Run(fmt.Sprintf("mkdir -p %s", target), p.Command.CommandTimeout)
	if err != nil {
		errChannel <- err
		return
	}
	if len(errStr) != 0 {
		errChannel <- fmt.Errorf(errStr)
		return
	}

	p.log("[artifact] untar file ", p.Artifact.tempFile)

	outStr, errStr, _, err := p.sshConfig.Run(fmt.Sprintf("tar -xf %s -C %s", p.Artifact.tempFile, target), p.Command.CommandTimeout)
	if p.Debug {
		p.log("[artifact] untar out ", outStr)
	}
	if err != nil {
		errChannel <- err
		return
	}
	if len(errStr) != 0 {
		p.log("[artifact] untar error: ", errStr)
	}
	for _, rp := range p.Artifact.RemovePrefix {
		if len(rp) == 0 {
			continue
		}
		p.log("[artifact] remove prefix: ", rp)
		outStr, errStr, _, err := p.sshConfig.Run(fmt.Sprintf("cd %s\nmv -f %s/* .\nrm -rf %s\n", target, rp, rp), p.Command.CommandTimeout)
		if len(errStr) != 0 {
			p.log("[artifact] remove prefix error: ", errStr)
		}
		if err != nil {
			errChannel <- err
			return
		}
		if p.Debug {
			p.log("[artifact] remove prefix out: ", outStr)
		}
	}
	wg.Done()

}

func (p *Plugin) removeTempFile() error {
	p.log("[artifact] remove temp all file/dir")
	_, errStr, _, err := p.sshConfig.Run(fmt.Sprintf("rm -rf %s", p.Artifact.tempFile), p.Command.CommandTimeout)
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return fmt.Errorf(errStr)
	}

	return nil
}
