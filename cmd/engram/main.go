package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "engram",
		Short:   "Engram — persistent, intelligent memory for AI coding agents",
		Version: version,
	}
	cmd.SetVersionTemplate(fmt.Sprintf("engram v%s\n", version))
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newIndexCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newRecallCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newTUICmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newImportCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
