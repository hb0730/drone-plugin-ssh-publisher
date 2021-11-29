package main

import (
	"strings"
	"sync"
)

func (p *Plugin) execCommand() error {
	if len(p.Command.Commands) == 0 {
		return nil
	}
	p.Command.Commands = p.ScriptCommands()
	p.log("======CMD======")
	p.log(strings.Join(p.Command.Commands, "\n"))
	p.log("======END======")
	wg := sync.WaitGroup{}
	wg.Add(1)
	errChannel := make(chan error)
	finished := make(chan struct{})
	go p.execCmd(strings.Join(p.Command.Commands, "\n"), errChannel, &wg)
	go func() {
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case err := <-errChannel:
		return err
	}
	p.log("==============================================")
	p.log("✅ Successfully executed commands to all host.")
	p.log("==============================================")
	return nil
}

func (p *Plugin) execCmd(cmd string, errChannel chan error, wg *sync.WaitGroup) {
	if len(cmd) == 0 {
		return
	}
	stdoutChan, stderrChan, doneChan, errChan, err := p.sshConfig.Stream(cmd, p.Command.CommandTimeout)
	if err != nil {
		errChannel <- err
	} else {
		isTimeout := true
	loop:
		for {
			select {
			case isTimeout = <-doneChan:
				break loop
			case outline := <-stdoutChan:
				if len(outline) > 0 {
					p.log(outline)
				}
			case errline := <-stderrChan:
				if len(errline) != 0 {
					p.log(errline)
				}
			case err = <-errChan:

			}
		}
		// get exit code or command error.
		if err != nil {
			errChannel <- err
		}
		// command time out
		if !isTimeout {
			errChannel <- errCommandTimeOut
		}
	}
	wg.Done()
}

func (p *Plugin) ScriptCommands() []string {
	cmds := make([]string, 0)
	for _, cmd := range p.Command.Commands {
		cmd = strings.TrimSpace(cmd)
		if len(strings.TrimSpace(cmd)) == 0 {
			continue
		}
		cmds = append(cmds, cmd)
	}
	return cmds
}
