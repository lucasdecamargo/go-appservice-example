package daemon

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/lucasdecamargo/kardianos"
)

type DaemonConfig struct {
	Executable string
	Args       []string
	EnvVars    []string

	OutWriter io.Writer
	ErrWriter io.Writer

	ExitTimeout time.Duration
}

type Daemon struct {
	DaemonConfig

	wg sync.WaitGroup

	cmd    *exec.Cmd
	retval error
}

func NewDaemon(cfg *DaemonConfig) *Daemon {
	return &Daemon{
		DaemonConfig: *cfg,
	}
}

func (p *Daemon) Start(s kardianos.Service) error {
	if p.Executable == "" {
		p.Executable, _ = os.Executable()
	}

	p.cmd = exec.Command(p.Executable, p.Args...)

	if len(p.EnvVars) > 0 {
		p.cmd.Env = append(os.Environ(), p.EnvVars...)
	}

	if p.OutWriter == nil {
		p.OutWriter = os.Stdout
	}

	if p.ErrWriter == nil {
		p.ErrWriter = os.Stderr
	}

	p.cmd.Stdout = p.OutWriter
	p.cmd.Stderr = p.ErrWriter

	p.wg.Add(1)

	go func() {
		defer func() {
			if !kardianos.Interactive() {
				s.Stop()
			} else {
				proc, _ := os.FindProcess(os.Getpid())
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					panic("failed signal termination to current process")
				}
			}
		}()

		defer p.wg.Done()

		p.retval = p.cmd.Run()
	}()

	return nil
}

func (p *Daemon) Stop(s kardianos.Service) error {
	if p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGTERM)
	}

	exit := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(exit)
	}()

	select {
	case <-exit:
		return p.retval
	case <-time.After(p.ExitTimeout):
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}

		return errors.New("program exit timeout")
	}
}
