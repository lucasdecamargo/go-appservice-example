package daemon

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/lucasdecamargo/kardianos"
)

const (
	defaultExitTimeout = 10 * time.Second
)

// DaemonConfig holds configuration for the daemon process supervisor
type DaemonConfig struct {
	Executable  string        // Path to the executable to run
	Args        []string      // Command line arguments
	EnvVars     []string      // Environment variables to set
	OutWriter   io.Writer     // Stdout writer
	ErrWriter   io.Writer     // Stderr writer
	ExitTimeout time.Duration // Timeout for graceful shutdown
}

// Daemon implements a process supervisor that can start, monitor, and stop child processes
type Daemon struct {
	DaemonConfig
	wg     sync.WaitGroup
	cmd    *exec.Cmd
	retval error
}

// NewDaemon creates a new daemon instance with the given configuration
func NewDaemon(cfg *DaemonConfig) *Daemon {
	if cfg.ExitTimeout == 0 {
		cfg.ExitTimeout = defaultExitTimeout
	}
	return &Daemon{DaemonConfig: *cfg}
}

// Start begins supervising the child process
func (d *Daemon) Start(s kardianos.Service) error {
	if d.Executable == "" {
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("executable path not found: %w", err)
		}
		d.Executable = executable
	}

	d.cmd = exec.Command(d.Executable, d.Args...)

	// Setup environment and IO
	if len(d.EnvVars) > 0 {
		d.cmd.Env = append(os.Environ(), d.EnvVars...)
	}
	if d.OutWriter == nil {
		d.OutWriter = os.Stdout
	}
	if d.ErrWriter == nil {
		d.ErrWriter = os.Stderr
	}
	d.cmd.Stdout = d.OutWriter
	d.cmd.Stderr = d.ErrWriter

	d.wg.Add(1)
	go d.superviseProcess(s)

	return nil
}

// Stop gracefully terminates the child process
func (d *Daemon) Stop(s kardianos.Service) error {
	if d.cmd.Process == nil {
		return nil
	}

	if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	return d.waitForProcessTermination()
}

// superviseProcess runs the child process and handles its lifecycle
func (d *Daemon) superviseProcess(s kardianos.Service) {
	defer func() {
		d.wg.Done()
		d.handleProcessExit(s)
	}()
	d.retval = d.cmd.Run()
}

// handleProcessExit manages what happens when the child process exits
func (d *Daemon) handleProcessExit(s kardianos.Service) {
	if !kardianos.Interactive() {
		s.Stop() // In service mode, stop the service when child exits
	} else {
		// In interactive mode, terminate the current process
		if proc, err := os.FindProcess(os.Getpid()); err == nil {
			proc.Signal(syscall.SIGTERM)
		}
	}
}

// waitForProcessTermination waits for the process to exit with timeout
func (d *Daemon) waitForProcessTermination() error {
	exit := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(exit)
	}()

	select {
	case <-exit:
		return d.retval
	case <-time.After(d.ExitTimeout):
		if d.cmd.Process != nil {
			d.cmd.Process.Kill()
		}
		return errors.New("program exit timeout")
	}
}
