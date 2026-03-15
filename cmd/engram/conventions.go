package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	conv "github.com/TomOst-Sec/colony-project/internal/conventions"
	"github.com/spf13/cobra"
)

func newConventionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conventions",
		Short: "Manage convention packs",
		Long:  "List, add, or remove community convention packs.",
	}
	cmd.AddCommand(newConventionsListCmd())
	cmd.AddCommand(newConventionsAddCmd())
	cmd.AddCommand(newConventionsRemoveCmd())
	return cmd
}

func getPacksDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".engram", "packs")
}

func newConventionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed convention packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := conv.NewPackRegistry(getPacksDir())
			packs, err := registry.List()
			if err != nil {
				return err
			}
			if len(packs) == 0 {
				fmt.Println("No convention packs installed.")
				fmt.Println("Install one: engram conventions add <pack-name>")
				return nil
			}
			for _, p := range packs {
				fmt.Printf("  %s v%s — %s (%d conventions)\n", p.Name, p.Version, p.Description, p.Count)
			}
			return nil
		},
	}
}

func newConventionsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <pack-name>",
		Short: "Download and install a convention pack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := conv.NewPackRegistry(getPacksDir())
			return registry.Install(context.Background(), args[0])
		},
	}
}

func newConventionsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <pack-name>",
		Short: "Remove an installed convention pack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := conv.NewPackRegistry(getPacksDir())
			if err := registry.Remove(args[0]); err != nil {
				return err
			}
			fmt.Printf("Removed '%s'\n", args[0])
			return nil
		},
	}
}
