package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var userTarget string

var userrolelistUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Print the list of users.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if userTarget != "all" && len(args) > 0 {
			return errors.New("can not specify arguments when you specified -t flag")
		}
		if len(args) > 1 {
			return errors.New("specify only one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var users *[]string = nil
		var err error = nil
		if len(args) > 0 {
			users, err = Client.ListUsersInRole(args[0])
		} else {
			switch userTarget {
			case "all":
				users, err = Client.ListUsers()
			case "permission":
				users, err = Client.ListPermissionUsers()
			default:
				return errors.New("Unsupported target: " + userTarget)
			}
		}
		if err != nil {
			return err
		}
		for _, user := range *users {
			fmt.Println(user)
		}
		return nil
	},
}

func init() {
	userrolelistCmd.AddCommand(userrolelistUsersCmd)
	userrolelistUsersCmd.PersistentFlags().StringVarP(&userTarget, "target", "t", "all", "Target roles.[all/permission]")
}
