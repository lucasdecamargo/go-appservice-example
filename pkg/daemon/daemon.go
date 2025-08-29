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
	// Default timeout for process termination
	defaultExitTimeout = 10 * time.Second

	// Error messages
	errExecutableNotFound = "executable path not found"
	errProcessTimeout     = "program exit timeout"
	errSignalTermination  = "failed to signal termination to current process"
)

// DaemonConfig holds configuration for the daemon process supervisor
type DaemonConfig struct {
	Executable string   // Path to the executable to run
	Args       []string // Command line arguments
	EnvVars    []string // Environment variables to set

	OutWriter io.Writer // Stdout writer
	ErrWriter io.Writer // Stderr writer

	ExitTimeout time.Duration // Timeout for graceful shutdown
}

// Daemon implements a process supervisor that can start, monitor, and stop child processes
type Daemon struct {
	DaemonConfig

	wg sync.WaitGroup

	cmd    *exec.Cmd
	retval error
}

// NewDaemon creates a new daemon instance with the given configuration
func NewDaemon(cfg *DaemonConfig) *Daemon {
	if cfg.ExitTimeout == 0 {
		cfg.ExitTimeout = defaultExitTimeout
	}

	return &Daemon{
		DaemonConfig: *cfg,
	}
}

// Start begins supervising the child process
func (d *Daemon) Start(s kardianos.Service) error {
	if err := d.setupExecutable(); err != nil {
		return fmt.Errorf("failed to setup executable: %w", err)
	}

	d.cmd = exec.Command(d.Executable, d.Args...)
	d.setupEnvironment()
	d.setupIO()

	d.wg.Add(1)

	go d.superviseProcess(s)

	return nil
}

// Stop gracefully terminates the child process
func (d *Daemon) Stop(s kardianos.Service) error {
	if d.cmd.Process == nil {
		return nil
	}

	// Send SIGTERM for graceful shutdown
	if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	return d.waitForProcessTermination()
}

// setupExecutable determines the executable path if not provided
func (d *Daemon) setupExecutable() error {
	if d.Executable == "" {
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("%s: %w", errExecutableNotFound, err)
		}
		d.Executable = executable
	}
	return nil
}

// setupEnvironment configures the process environment
func (d *Daemon) setupEnvironment() {
	if len(d.EnvVars) > 0 {
		d.cmd.Env = append(os.Environ(), d.EnvVars...)
	}
}

// setupIO configures input/output streams
func (d *Daemon) setupIO() {
	if d.OutWriter == nil {
		d.OutWriter = os.Stdout
	}
	if d.ErrWriter == nil {
		d.ErrWriter = os.Stderr
	}

	d.cmd.Stdout = d.OutWriter
	d.cmd.Stderr = d.ErrWriter
}

// superviseProcess runs the child process and handles its lifecycle
func (d *Daemon) superviseProcess(s kardianos.Service) {
	defer func() {
		d.handleProcessExit(s)
		d.wg.Done()
	}()

	d.retval = d.cmd.Run()
}

// handleProcessExit manages what happens when the child process exits
func (d *Daemon) handleProcessExit(s kardianos.Service) {
	if !kardianos.Interactive() {
		// In service mode, stop the service when child exits
		s.Stop()
	} else {
		// In interactive mode, terminate the current process
		proc, err := os.FindProcess(os.Getpid())
		if err != nil {
			panic(fmt.Sprintf("%s: %v", errSignalTermination, err))
		}

		if err := proc.Signal(syscall.SIGTERM); err != nil {
			panic(fmt.Sprintf("%s: %v", errSignalTermination, err))
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
		// Force kill if timeout exceeded
		if d.cmd.Process != nil {
			if err := d.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}
		return errors.New(errProcessTimeout)
	}
}
