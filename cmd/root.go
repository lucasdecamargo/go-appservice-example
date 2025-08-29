package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "svcapp",
		Short: "A simple example of a Go application that can be installed as a service",
	}
}
