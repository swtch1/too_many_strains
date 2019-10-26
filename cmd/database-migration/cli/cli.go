package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	Help             bool
	LogLevel         string
	DatabaseUsername string
	DatabasePassword string
	SeedFile         string
)

// Init performs setup for the application CLI commands and flags, setting application version as provided.
func Init(appName, version string) {
	// cmd is the root of our CLI
	cmd := &cobra.Command{
		Use:   appName,
		Short: appName,
		Long:  fmt.Sprintf("%s is a database migration script which will ensure the strain server database is on the correct schema version", appName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.PersistentFlags().BoolVarP(&Help, "help", "h", false, "Display this help and exit.")
	cmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "l", "info", "Log level should be one of trace, debug, info, warn, error, fatal.")
	cmd.PersistentFlags().StringVarP(&DatabaseUsername, "database-username", "u", "root", "The username of the database.")
	cmd.PersistentFlags().StringVarP(&DatabasePassword, "database-password", "p", "password", "The password of the database.")
	cmd.PersistentFlags().StringVarP(&SeedFile, "database-seed-file", "f", "./strains.json", "Path to JSON strains file which will seed the database.")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if Help {
		os.Exit(0)
	}
}
