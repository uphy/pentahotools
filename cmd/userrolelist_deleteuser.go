package cmd

import (
	"github.com/spf13/cobra"
)

var userrolelistDeleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete the specified users",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.DeleteUser(args...)
	},
}

func init() {
	userrolelistCmd.AddCommand(userrolelistDeleteUserCmd)
}
