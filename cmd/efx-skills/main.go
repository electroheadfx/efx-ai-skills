package main

import (
	"fmt"
	"os"

	"github.com/lmarques/efx-skills/internal/tui"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "efx-skills",
		Short:   "Unified AI agent skills manager",
		Long:    `efx-skills is a TUI tool for discovering, previewing, installing, and managing AI agent skills across multiple providers.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}

	// Search command
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search skills from skills.sh and playbooks.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			return tui.RunSearch(query)
		},
	}

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show provider status panel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunStatus()
		},
	}

	// Preview command
	previewCmd := &cobra.Command{
		Use:   "preview <skill>",
		Short: "Preview skill SKILL.md content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunPreview(args[0])
		},
	}

	// Install command
	installCmd := &cobra.Command{
		Use:   "install <skill>",
		Short: "Install skill to selected providers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providers, _ := cmd.Flags().GetStringSlice("provider")
			return tui.RunInstall(args[0], providers)
		},
	}
	installCmd.Flags().StringSliceP("provider", "p", []string{}, "Target providers (claude, cursor, qoder, etc.)")

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunList()
		},
	}

	// Sync command
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync skills across all providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunSync()
		},
	}

	// Config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration and custom sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunConfig()
		},
	}

	rootCmd.AddCommand(searchCmd, statusCmd, previewCmd, installCmd, listCmd, syncCmd, configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
