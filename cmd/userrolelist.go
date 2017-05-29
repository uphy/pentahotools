package cmd

import (
	"errors"
	"fmt"

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

var userrolelistChangeUserPasswordCmd = &cobra.Command{
	Use:   "change-user-password",
	Short: "Change the user password.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return errors.New("Specify 3 arguments(username,oldpassword,newpassword)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.ChangeUserPassword(args[0], args[1], args[2])
	},
}

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

var userrolelistDeleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete the specified users",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.DeleteUser(args...)
	},
}

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
	RootCmd.AddCommand(userrolelistCmd)

	userrolelistCmd.AddCommand(userrolelistChangeUserPasswordCmd)

	userrolelistCmd.AddCommand(userrolelistCreateUserCmd)

	userrolelistCmd.AddCommand(userrolelistDeleteUserCmd)

	userrolelistrolesCmd.PersistentFlags().StringVarP(&roleTarget, "target", "t", "all", "Target roles.[all/standard/permission/system/extra]")
	userrolelistCmd.AddCommand(userrolelistrolesCmd)

	userrolelistUsersCmd.PersistentFlags().StringVarP(&userTarget, "target", "t", "all", "Target roles.[all/permission]")
	userrolelistCmd.AddCommand(userrolelistUsersCmd)
}
