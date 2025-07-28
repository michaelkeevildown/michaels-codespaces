package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/michaelkeevildown/mcs/internal/backup"
	"github.com/spf13/cobra"
)

// BackupCommand creates the 'backup' command
func BackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "💾 Manage MCS backups",
		Long:  `Create, list, restore, and manage backups of your codespaces and MCS installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		backupCreateCommand(),
		backupListCommand(),
		backupRestoreCommand(),
		backupDeleteCommand(),
		backupCleanupCommand(),
	)

	return cmd
}

// backupCreateCommand creates the 'backup create' subcommand
func backupCreateCommand() *cobra.Command {
	var (
		backupType  string
		description string
		sourcePath  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new backup",
		Long:  `Create a backup of your codespaces or MCS installation.`,
		Example: `  # Backup codespaces
  mcs backup create --type manual

  # Backup a specific directory
  mcs backup create --source /path/to/directory --description "Before major update"

  # Backup MCS installation
  mcs backup create --type install --source ~/.mcs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backupManager := backup.NewBackupManager()

			// Determine source path
			if sourcePath == "" {
				homeDir := os.Getenv("HOME")
				switch backupType {
				case "install":
					sourcePath = filepath.Join(homeDir, ".mcs")
				default:
					sourcePath = filepath.Join(homeDir, "codespaces")
				}
			}

			// Validate backup type
			var bType backup.BackupType
			switch backupType {
			case "destroy":
				bType = backup.BackupTypeDestroy
			case "install":
				bType = backup.BackupTypeInstall
			case "manual":
				bType = backup.BackupTypeManual
			default:
				return fmt.Errorf("invalid backup type: %s (must be: destroy, install, or manual)", backupType)
			}

			// Check if source exists
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				return fmt.Errorf("source path does not exist: %s", sourcePath)
			}

			fmt.Println(infoStyle.Render("📦 Creating Backup"))
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("Source: %s\n", sourcePath)
			fmt.Printf("Type: %s\n", backupType)
			if description != "" {
				fmt.Printf("Description: %s\n", description)
			}
			fmt.Println()

			// Create backup
			backupID, err := backupManager.Create(sourcePath, bType, description)
			if err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}

			fmt.Println(successStyle.Render("✓ Backup created successfully!"))
			fmt.Printf("Backup ID: %s\n", boldStyle.Render(backupID))
			fmt.Printf("Location: %s\n", dimStyle.Render("~/.mcs.backup/"+backupID))
			fmt.Println()
			fmt.Println(dimStyle.Render("Use 'mcs backup restore " + backupID + "' to restore this backup"))

			return nil
		},
	}

	cmd.Flags().StringVarP(&backupType, "type", "t", "manual", "Backup type (destroy, install, manual)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Backup description")
	cmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source directory to backup")

	return cmd
}

// backupListCommand creates the 'backup list' subcommand
func backupListCommand() *cobra.Command {
	var (
		showDetails bool
		filterType  string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all backups",
		Long:    `Display all available backups with their metadata.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backupManager := backup.NewBackupManager()

			backups, err := backupManager.List()
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) == 0 {
				fmt.Println(dimStyle.Render("No backups found"))
				return nil
			}

			fmt.Println(headerStyle.Render("💾 Available Backups"))
			fmt.Println(strings.Repeat("═", 80))
			fmt.Println()

			if showDetails {
				// Detailed view
				for i, b := range backups {
					if filterType != "" && string(b.Type) != filterType {
						continue
					}

					fmt.Printf("%s %s\n", boldStyle.Render("Backup ID:"), b.ID)
					fmt.Printf("%s %s\n", dimStyle.Render("Type:"), string(b.Type))
					fmt.Printf("%s %s\n", dimStyle.Render("Created:"), b.Timestamp.Format("2006-01-02 15:04:05"))
					fmt.Printf("%s %s\n", dimStyle.Render("Size:"), backup.FormatSize(b.Size))
					fmt.Printf("%s %s\n", dimStyle.Render("Source:"), b.SourcePath)
					if b.Description != "" {
						fmt.Printf("%s %s\n", dimStyle.Render("Description:"), b.Description)
					}
					
					if i < len(backups)-1 {
						fmt.Println(strings.Repeat("─", 50))
					}
				}
			} else {
				// Table view
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "ID\tTYPE\tCREATED\tSIZE\tDESCRIPTION")
				fmt.Fprintln(w, strings.Repeat("─", 70))

				for _, b := range backups {
					if filterType != "" && string(b.Type) != filterType {
						continue
					}

					desc := b.Description
					if desc == "" {
						desc = "-"
					}
					if len(desc) > 30 {
						desc = desc[:27] + "..."
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						b.ID,
						string(b.Type),
						b.Timestamp.Format("2006-01-02 15:04"),
						backup.FormatSize(b.Size),
						desc,
					)
				}
				w.Flush()
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information")
	cmd.Flags().StringVarP(&filterType, "type", "t", "", "Filter by backup type")

	return cmd
}

// backupRestoreCommand creates the 'backup restore' subcommand
func backupRestoreCommand() *cobra.Command {
	var (
		targetPath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "restore [backup-id]",
		Short: "Restore a backup",
		Long:  `Restore a previously created backup to a specified location.`,
		Args:  cobra.ExactArgs(1),
		Example: `  # Restore to original location
  mcs backup restore destroy-20250728_123456

  # Restore to custom location
  mcs backup restore destroy-20250728_123456 --target /tmp/restored`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backupID := args[0]
			backupManager := backup.NewBackupManager()

			// Check if backup exists
			if !backupManager.Exists(backupID) {
				return fmt.Errorf("backup not found: %s", backupID)
			}

			// Get backup info
			backups, err := backupManager.List()
			if err != nil {
				return fmt.Errorf("failed to get backup info: %w", err)
			}

			var backupInfo *backup.BackupInfo
			for _, b := range backups {
				if b.ID == backupID {
					backupInfo = &b
					break
				}
			}

			if backupInfo == nil {
				return fmt.Errorf("backup metadata not found: %s", backupID)
			}

			// Determine target path
			if targetPath == "" {
				// Use parent directory of original source
				targetPath = filepath.Dir(backupInfo.SourcePath)
			}

			fmt.Println(infoStyle.Render("📦 Restoring Backup"))
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("Backup ID: %s\n", backupID)
			fmt.Printf("Type: %s\n", string(backupInfo.Type))
			fmt.Printf("Created: %s\n", backupInfo.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("Target: %s\n", targetPath)
			fmt.Println()

			// Check if target already exists
			targetExists := false
			entries, _ := os.ReadDir(backupManager.GetBackupPath(backupID))
			for _, entry := range entries {
				if entry.Name() == "metadata.json" {
					continue
				}
				destPath := filepath.Join(targetPath, entry.Name())
				if _, err := os.Stat(destPath); err == nil {
					targetExists = true
					break
				}
			}

			if targetExists && !force {
				fmt.Println(warningStyle.Render("⚠️  Target location already contains data"))
				fmt.Println("Use --force to overwrite existing data")
				return fmt.Errorf("target already exists")
			}

			// Perform restore
			if err := backupManager.Restore(backupID, targetPath); err != nil {
				return fmt.Errorf("failed to restore backup: %w", err)
			}

			fmt.Println(successStyle.Render("✓ Backup restored successfully!"))
			fmt.Printf("Restored to: %s\n", boldStyle.Render(targetPath))

			return nil
		},
	}

	cmd.Flags().StringVarP(&targetPath, "target", "t", "", "Target directory for restoration")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force restore even if target exists")

	return cmd
}

// backupDeleteCommand creates the 'backup delete' subcommand
func backupDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete [backup-id]",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a backup",
		Long:    `Permanently delete a backup.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupID := args[0]
			backupManager := backup.NewBackupManager()

			// Check if backup exists
			if !backupManager.Exists(backupID) {
				return fmt.Errorf("backup not found: %s", backupID)
			}

			if !force {
				fmt.Printf("Are you sure you want to delete backup '%s'? (y/N): ", backupID)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Delete backup
			if err := backupManager.Delete(backupID); err != nil {
				return fmt.Errorf("failed to delete backup: %w", err)
			}

			fmt.Println(successStyle.Render("✓ Backup deleted successfully"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// backupCleanupCommand creates the 'backup cleanup' subcommand
func backupCleanupCommand() *cobra.Command {
	var (
		keepCount int
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old backups",
		Long:  `Remove old backups, keeping only the specified number of most recent backups.`,
		Example: `  # Keep only the 5 most recent backups
  mcs backup cleanup --keep 5

  # Preview what would be deleted
  mcs backup cleanup --keep 5 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backupManager := backup.NewBackupManager()

			// List current backups
			backups, err := backupManager.List()
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) <= keepCount {
				fmt.Printf("Currently have %d backups, keeping %d. Nothing to clean up.\n", 
					len(backups), keepCount)
				return nil
			}

			fmt.Println(infoStyle.Render("🧹 Backup Cleanup"))
			fmt.Println(strings.Repeat("─", 50))
			fmt.Printf("Total backups: %d\n", len(backups))
			fmt.Printf("Keeping: %d most recent\n", keepCount)
			fmt.Printf("Deleting: %d old backups\n", len(backups)-keepCount)
			fmt.Println()

			// Show which backups will be deleted
			fmt.Println("Backups to be deleted:")
			for i := keepCount; i < len(backups); i++ {
				age := time.Since(backups[i].Timestamp)
				fmt.Printf("  • %s (%s old, %s)\n", 
					backups[i].ID, 
					formatBackupAge(age),
					backup.FormatSize(backups[i].Size))
			}

			if dryRun {
				fmt.Println()
				fmt.Println(dimStyle.Render("(Dry run - no changes made)"))
				return nil
			}

			fmt.Println()
			// Perform cleanup
			if err := backupManager.CleanupOld(keepCount); err != nil {
				return fmt.Errorf("failed to cleanup backups: %w", err)
			}

			fmt.Println(successStyle.Render("✓ Cleanup completed successfully"))
			return nil
		},
	}

	cmd.Flags().IntVarP(&keepCount, "keep", "k", 5, "Number of recent backups to keep")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be deleted without making changes")

	return cmd
}

// formatBackupAge formats a duration in a human-readable way for backups
func formatBackupAge(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d days", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return "just now"
}