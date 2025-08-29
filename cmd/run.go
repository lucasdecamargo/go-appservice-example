package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	signalBufferSize = 3
	shutdownTimeout  = 60 * time.Second
)

// RunFunc represents the function signature for the main application logic
type RunFunc func(ctx context.Context, args []string) error

// NewRunCmd creates a command for running the application with signal handling
func NewRunCmd(f RunFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the application and exit with the specified status",
		Long: `Run the application with signal handling and graceful shutdown.

The run command executes the application with proper signal handling for SIGINT (Ctrl+C) 
and SIGTERM. It ensures graceful shutdown by canceling the context and waiting for 
the application to complete.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithSignals(cmd.Context(), f, args)
		},
	}
}

// runWithSignals executes the application with signal handling
func runWithSignals(ctx context.Context, f RunFunc, args []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, signalBufferSize)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Run application in goroutine
	done := make(chan struct{})
	var runErr error
	var wg sync.WaitGroup

	wg.Go(func() {
		defer close(done)
		runErr = f(ctx, args)
	})

	// Wait for completion or signal
	select {
	case <-done:
		return runErr
	case sig := <-sigChan:
		return handleShutdown(cancel, &wg, sig, runErr)
	}
}

// handleShutdown manages graceful shutdown
func handleShutdown(cancel context.CancelFunc, wg *sync.WaitGroup, sig os.Signal, runErr error) error {
	cancel()

	shutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		if runErr != nil {
			return fmt.Errorf("application error: %w", runErr)
		}
		return fmt.Errorf("shutdown by signal: %v", sig)
	case <-time.After(shutdownTimeout):
		return fmt.Errorf("shutdown timeout exceeded after %v", shutdownTimeout)
	}
}
