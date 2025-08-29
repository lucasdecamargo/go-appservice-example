package cmd

import (
	"fmt"
	"os"

	"github.com/lucasdecamargo/go-appservice-example/pkg/daemon"
	"github.com/lucasdecamargo/kardianos"
	"github.com/spf13/cobra"
)

// NewDaemonCmd creates a command for running the application as a daemon process supervisor.
// The daemon command runs the application in service mode, supervising child processes
// and managing their lifecycle. It requires root privileges for proper service operation.
//
// The daemon command:
// - Runs the application as a service using the kardianos service framework
// - Supervises child processes and restarts them on failure
// - Handles graceful shutdowns and signal management
// - Supports additional command-line arguments passed to the child process
//
// Usage:
//
//	svcapp daemon                    # Run with default configuration
//	svcapp daemon -v --flag val      # Run with additional arguments
//	sudo svcapp daemon               # Run with root privileges (recommended)
//
// Parameters:
//
//	d:   The daemon instance that implements process supervision
//	cfg: Service configuration for the target operating system
//
// Returns:
//
//	A configured cobra.Command that handles daemon execution
func NewDaemonCmd(d *daemon.Daemon, cfg *kardianos.Config) *cobra.Command {
	c := &cobra.Command{
		Use:                "daemon",
		Short:              "Manage the daemon service. Requires root privileges.",
		Long:               "Run the application as a daemon process supervisor that monitors and restarts child processes.",
		DisableFlagParsing: true, // Allow passing arbitrary arguments to child process
		Run: func(cmd *cobra.Command, args []string) {
			// Append any additional arguments to the daemon's argument list
			if len(args) > 0 {
				d.Args = append(d.Args, args...)
			}

			// Create and start the service
			s, err := kardianos.New(d, cfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Run the service (this blocks until the service stops)
			if err := s.Run(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	return c
}
