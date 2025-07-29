package cli

import (
	"fmt"

	"github.com/michaelkeevildown/mcs/internal/version"
	"github.com/spf13/cobra"
)

// VersionCommand returns the version command
func VersionCommand() *cobra.Command {
	var detailed bool
	
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  "Display version information for MCS",
		Run: func(cmd *cobra.Command, args []string) {
			if detailed {
				fmt.Println(version.DetailedInfo())
			} else {
				fmt.Printf("mcs version %s\n", version.Info())
			}
		},
	}
	
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed version information")
	
	return cmd
}