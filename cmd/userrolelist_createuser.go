package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var userrolelistCreateUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create user",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Specify 2 arguments(username and password)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.CreateUser(args[0], args[1])
	},
}

func init() {
	userrolelistCmd.AddCommand(userrolelistCreateUserCmd)
}
