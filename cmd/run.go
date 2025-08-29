package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

func NewRunCmd(f func(ctx context.Context, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the application and exit with the specified status",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, cancel := context.WithCancel(cmd.Context())

			// Set up channel to listen for interrupt (Ctrl+C)
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			wg := sync.WaitGroup{}
			done := make(chan struct{})

			wg.Go(func() {
				err = f(ctx, args) // Set the return value
				close(done)
			})

			select {
			case <-done:
			case <-sigChan:
			}

			cancel()

			wg.Wait()

			return
		},
	}
}
