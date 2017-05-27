package cmd

import (
	"fmt"

	"errors"

	"github.com/spf13/cobra"
)

var roleTarget string

var userrolelistrolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Print the list of roles.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if roleTarget != "all" && len(args) > 0 {
			return errors.New("can not specify arguments when you specified -t flag")
		}
		if len(args) > 1 {
			return errors.New("specify only one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var roles *[]string = nil
		var err error = nil
		if len(args) > 0 {
			roles, err = Client.ListRolesForUser(args[0])
		} else {
			switch roleTarget {
			case "all":
				roles, err = Client.ListAllRoles()
			case "standard":
				roles, err = Client.ListRoles()
			case "permission":
				roles, err = Client.ListPermissionRoles()
			case "system":
				roles, err = Client.ListSystemRoles()
			case "extra":
				roles, err = Client.ListExtraRoles()
			default:
				return fmt.Errorf("Unsupported role list type: %s", roleTarget)
			}
		}
		if err != nil {
			return err
		}
		for _, role := range *roles {
			fmt.Println(role)
		}
		return nil
	},
}

func init() {
	userrolelistCmd.AddCommand(userrolelistrolesCmd)
	userrolelistrolesCmd.PersistentFlags().StringVarP(&roleTarget, "target", "t", "all", "Target roles.[all/standard/permission/system/extra]")
}
