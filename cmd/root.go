package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "svcapp",
		Short: "A brief description of your application",
	}
}
