package cmd

import (
	"fmt"
	"os"

	"github.com/lucasdecamargo/go-appservice-example/pkg/daemon"
	"github.com/lucasdecamargo/kardianos"
	"github.com/spf13/cobra"
)

func NewDaemonCmd(d *daemon.Daemon, cfg *kardianos.Config) *cobra.Command {
	c := &cobra.Command{
		Use:                "daemon",
		Short:              "Manage the daemon service. Requires root privileges.",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				d.Args = append(d.Args, args...)
			}

			s, err := kardianos.New(d, cfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err := s.Run(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	return c
}
