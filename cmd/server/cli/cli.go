package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	Help                bool
	Version             bool
	Port                int32
	DatabaseUsername    string
	DatabasePassword    string
	DatabaseName        string
	LogLevel            string
	LogFormat           string
	PrettyPrintJsonLogs bool
)

// Init performs setup for the application CLI commands and flags, setting application version as provided.
func Init(appName, version string) {
	// cmd is the root of our CLI
	cmd := &cobra.Command{
		Use:   appName,
		Short: appName,
		Long:  fmt.Sprintf("%s is a Cannabis strains server.", appName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.PersistentFlags().BoolVarP(&Help, "help", "h", false, "Display this help and exit.")
	cmd.PersistentFlags().BoolVar(&Version, "version", false, "Print the application version and exit.")
	cmd.PersistentFlags().Int32VarP(&Port, "port", "P", 8888, "Port which the server will listen on.")
	cmd.PersistentFlags().StringVarP(&DatabaseUsername, "db-username", "u", "root", "Database username.")
	cmd.PersistentFlags().StringVarP(&DatabasePassword, "db-password", "p", "password", "Database password.")
	cmd.PersistentFlags().StringVar(&DatabaseName, "db-name", "so_many_strains", "Name of the logical database.")
	cmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "l", "info", "Log level should be one of trace, debug, info, warn, error, fatal.")
	cmd.PersistentFlags().StringVar(&LogFormat, "log-format", "text", "Log format should be one of text, json.")
	cmd.PersistentFlags().BoolVar(&PrettyPrintJsonLogs, "pretty-json", false, "If writing JSON logs, pretty print those logs.")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if Help {
		os.Exit(0)
	}

	// handle the version manually since the built in version options for Cobra do not exit after printing
	if Version {
		fmt.Printf("%s version %s\n", appName, version)
		os.Exit(0)
	}
}
