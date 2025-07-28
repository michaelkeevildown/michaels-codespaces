package main

import (
	"fmt"
	"os"

	"github.com/michaelkeevildown/mcs/internal/cli"
	"github.com/michaelkeevildown/mcs/internal/update"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcs",
		Short: "🚀 Michael's Codespaces - AI-powered development environments",
		Long: `Michael's Codespaces (MCS) provides isolated, reproducible development
environments optimized for AI agents and modern development workflows.

Run AI agents without constraints, on your own hardware.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip update check for certain commands
			skipCommands := map[string]bool{
				"update":     true,
				"autoupdate": true,
				"version":    true,
				"help":       true,
				"completion": true,
			}
			
			if !skipCommands[cmd.Name()] {
				// Check for updates in the background
				go update.CheckForUpdates(version)
			}
		},
	}

	// Add commands
	rootCmd.AddCommand(
		cli.SetupCommand(),
		cli.CreateCommand(),
		cli.ListCommand(),
		cli.StartCommand(),
		cli.StopCommand(),
		cli.RestartCommand(),
		cli.RebuildCommand(),
		cli.RemoveCommand(),
		cli.ExecCommand(),
		cli.LogsCommand(),
		cli.InfoCommand(),
		cli.RecoverCommand(),
		cli.ResetPasswordCommand(),
		cli.UpdateIPCommand(),
		cli.AutoUpdateCommand(),
		cli.DoctorCommand(),
		cli.StatusCommand(),
		cli.UpdateCommand(),
		cli.CleanupCommand(),
		cli.DestroyCommand(),
		cli.BackupCommand(),
	)

	// Customize help
	rootCmd.SetHelpTemplate(helpTemplate())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func helpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
}