package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0-dev"
	commit  = "none"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "engram",
		Short:   "Engram — persistent, intelligent memory for AI coding agents",
		Version: version,
	}
	cmd.SetVersionTemplate(fmt.Sprintf("engram v%s (%s)\n", version, commit))
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newIndexCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newRecallCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newTUICmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newCIHookCmd())
	cmd.AddCommand(newConventionsCmd())
	cmd.AddCommand(newCallgraphCmd())
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
