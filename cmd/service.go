package cmd

import (
	"fmt"
	"os"

	"github.com/lucasdecamargo/kardianos"
	"github.com/spf13/cobra"
)

func NewServiceCmd(i kardianos.Interface, cfg *kardianos.Config) *cobra.Command {
	return &cobra.Command{
		Use:       "service {start|stop|restart|install|uninstall}",
		Short:     "Manage the application service. Requires root privileges.",
		ValidArgs: []string{"start", "stop", "restart", "install", "uninstall"},
		Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			s, err := kardianos.New(i, cfg)
			if err != nil {
				panic(err)
			}

			if err := kardianos.Control(s, args[0]); err != nil {
				switch err {
				case kardianos.ErrNotInstalled:
					fmt.Println("Error: Service is not installed. Run 'install' to install it.")
					os.Exit(1)
				case kardianos.ErrNoServiceSystemDetected:
					fmt.Println("Error: Could not detect service system.")
					os.Exit(1)
				case kardianos.ErrServiceExists:
					fmt.Println("Already installed.")
				default:
					fmt.Println(err)
					fmt.Println("Obs: Service commands require root privileges.")
					os.Exit(1)
				}
			}
		},
	}
}
