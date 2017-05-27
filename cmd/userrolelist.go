package cmd

import (
	"github.com/spf13/cobra"
)

var userrolelistCmd = &cobra.Command{
	Use:   "userrolelist",
	Short: "User management command",
	Long:  `Manage the users of Pentaho.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(userrolelistCmd)
}
