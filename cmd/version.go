package cmd

import (
	"fmt"
	"github.com/osodracnai/go-basic-server/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the application version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", version.Version)
		fmt.Println("GitCommit:", version.GitCommit)
		fmt.Println("BuildTime:", version.BuildTime)
		fmt.Println("OS/Arch:", fmt.Sprintf("%s/%s", version.OS, version.Arch))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
