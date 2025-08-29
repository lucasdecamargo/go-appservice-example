package cmd

import (
	"fmt"
	"os"

	"github.com/lucasdecamargo/kardianos"
	"github.com/spf13/cobra"
)

const (
	// Error messages
	errServiceNotInstalled = "Error: Service is not installed. Run 'install' to install it."
	errNoServiceSystem     = "Error: Could not detect service system."
	errAlreadyInstalled    = "Already installed."
)

// NewServiceCmd creates a command for managing the application service
func NewServiceCmd(i kardianos.Interface, cfg *kardianos.Config) *cobra.Command {
	return &cobra.Command{
		Use:       "service {start|stop|restart|install|uninstall}",
		Short:     "Manage the application service. Requires root privileges.",
		ValidArgs: []string{"start", "stop", "restart", "install", "uninstall"},
		Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			if err := handleServiceCommand(i, cfg, args[0]); err != nil {
				os.Exit(1)
			}
		},
	}
}

// handleServiceCommand processes service management commands
func handleServiceCommand(i kardianos.Interface, cfg *kardianos.Config, action string) error {
	s, err := kardianos.New(i, cfg)
	if err != nil {
		panic(err) // not supposed to happen in production
	}

	if err := kardianos.Control(s, action); err != nil {
		return handleServiceError(err)
	}

	return nil
}

// handleServiceError processes service-related errors and provides user-friendly messages
func handleServiceError(err error) error {
	switch err {
	case kardianos.ErrNotInstalled:
		fmt.Println(errServiceNotInstalled)
	case kardianos.ErrNoServiceSystemDetected:
		fmt.Println(errNoServiceSystem)
	case kardianos.ErrServiceExists:
		fmt.Println(errAlreadyInstalled)
		return nil // Not an error, just informational
	default:
		fmt.Printf("Service error: %v\n", err)
	}

	return err
}
