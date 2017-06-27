package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/uphy/pentahotools/table"
	"gopkg.in/cheggaaa/pb.v1"
)

var userroleCmd = &cobra.Command{
	Use:   "userrole",
	Short: "User management command",
	Long:  `Manage the users of Pentaho.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var userroleChangeUserPasswordCmd = &cobra.Command{
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

var userroleUpdatePasswordCmd = &cobra.Command{
	Use:   "update-password",
	Short: "Change the user password.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Specify 2 arguments(username,newpassword)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.UpdatePassword(args[0], args[1])
	},
}

var userroleCreateRoleCmd = &cobra.Command{
	Use:   "create-role",
	Short: "Create role",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("specify at least 1 role")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, role := range args {
			err := Client.CreateRole(role)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

var userroleDeleteRoleCmd = &cobra.Command{
	Use:   "delete-role",
	Short: "Delete roles",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("specify at least 1 role")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Client.DeleteRoles(args...)
	},
}

var userroleListPermissionsCmd = &cobra.Command{
	Use:   "list-permissions",
	Short: "List the permissions for the role.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var roles *[]Role
		var err error
		switch len(args) {
		case 0:
			roles, err = NewRoleClient(&Client).FindAllRoles()
		case 1:
			var role *Role
			role, err = NewRoleClient(&Client).FindRole(args[0])
			roles = &([]Role{*role})
		}
		if err != nil {
			return errors.Wrap(err, "Failed to list the permissions for roles")
		}
		for _, role := range *roles {
			role.Print()
			fmt.Println()
		}
		return nil
	},
}

var userroleAssignPermissionsCmd = &cobra.Command{
	Use:   "set-permissions",
	Short: "Set the permissions of the role.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("specify role name and permissions")
		}
		var err error
		if len(args) == 1 {
			err = NewRoleClient(&Client).SetPermissionsOfRole(args[0])
		} else {
			err = NewRoleClient(&Client).SetPermissionsOfRole(args[0], args[1:]...)
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to set the permissions of the role. (role=%s, permissions=%s)", args[0], args[1:]))
		}
		return nil
	},
}

var userroleAddPermissionsCmd = &cobra.Command{
	Use:   "add-permissions",
	Short: "Add the permissions to the role.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("specify role name and permissions")
		}
		err := NewRoleClient(&Client).AddPermissionsToRole(args[0], args[1:]...)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to add the permissions to the role. (role=%s, permissions=%s)", args[0], args[1:]))
		}
		return nil
	},
}

var userroleRemovePermissionsCmd = &cobra.Command{
	Use:   "remove-permissions",
	Short: "Remove the permissions from the role.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("specify role name and permissions")
		}
		err := NewRoleClient(&Client).RemovePermissionsFromRole(args[0], args[1:]...)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to add the permissions to the role. (role=%s, permissions=%s)", args[0], args[1:]))
		}
		return nil
	},
}

var file string

var userroleCreateUserCmd = &cobra.Command{
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

var userroleDeleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete the specified users",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := cmd.Flags().GetBool("homeDir")

		bar := pb.StartNew(0)
		err := DeleteUsers(args, homeDir, bar)
		bar.FinishPrint("Finished to delete users.")
		return err
	},
}

var userroleImportUsersCmd = &cobra.Command{
	Use:   "import-users",
	Short: "Import users.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("specify a users file(csv/xlsx)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		bar := pb.StartNew(0)
		deleteUsers, _ := cmd.Flags().GetBool("delete-users")
		deleteHomeDir, _ := cmd.Flags().GetBool("delete-homedir")
		updatePassword, _ := cmd.Flags().GetBool("update-password")
		defaultPassword, _ := cmd.Flags().GetString("default-password")
		headerSize, _ := cmd.Flags().GetInt("header-size")
		err := ImportUsers(args[0], &ImportUsersOptions{
			DeleteUsers:         deleteUsers,
			DeleteHomeDirectory: deleteHomeDir,
			UpdatePassword:      updatePassword,
			DefaultPassword:     defaultPassword,
			HeaderSize:          headerSize,
			Separator:           "\t",
		}, bar)
		bar.FinishPrint("Finished to create the users.")
		return err
	},
}

var userroleExportUsersCmd = &cobra.Command{
	Use:   "export-users",
	Short: "Export all of the users",
	RunE: func(cmd *cobra.Command, args []string) error {
		var file string
		switch len(args) {
		case 0:
			file = table.ConsoleOutput
		case 1:
			file = args[0]
		default:
			return errors.New("specify a file")
		}
		headers, _ := cmd.Flags().GetBool("headers")
		bar := pb.StartNew(0)
		err := ExportUsers(file, &ExportUsersOptions{
			Separator:  "\t",
			WithHeader: headers,
		}, bar)
		bar.FinishPrint("Finished to export users.")
		return err
	},
}

var roleTarget string

var userrolerolesCmd = &cobra.Command{
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

var userroleUsersCmd = &cobra.Command{
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

var userroleAssignRoleToUserCmd = &cobra.Command{
	Use:   "assign-role-to-user",
	Short: "Assigns the roles to a user.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("specify username and roles")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		userName := args[0]
		roles := args[1:]
		return Client.AssignRolesToUser(userName, roles...)
	},
}

var userroleRemoveRoleFromUserCmd = &cobra.Command{
	Use:   "remove-role-from-user",
	Short: "Removes the roles from a user.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("specify username and roles")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		userName := args[0]
		roles := args[1:]
		return Client.RemoveRolesFromUser(userName, roles...)
	},
}

func init() {
	RootCmd.AddCommand(userroleCmd)

	userroleCmd.AddCommand(userroleChangeUserPasswordCmd)

	userroleCmd.AddCommand(userroleUpdatePasswordCmd)

	userroleCmd.AddCommand(userroleCreateRoleCmd)
	userroleCmd.AddCommand(userroleDeleteRoleCmd)
	userroleCmd.AddCommand(userroleListPermissionsCmd)
	userroleCmd.AddCommand(userroleAssignPermissionsCmd)
	userroleCmd.AddCommand(userroleAddPermissionsCmd)
	userroleCmd.AddCommand(userroleRemovePermissionsCmd)

	userroleCreateUserCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "Batch create from CSV file.")
	userroleCmd.AddCommand(userroleCreateUserCmd)

	userroleDeleteUserCmd.Flags().BoolP("homeDir", "H", false, "Also delete home directory.")
	userroleCmd.AddCommand(userroleDeleteUserCmd)

	userroleExportUsersCmd.Flags().BoolP("headers", "e", false, "Print headers.")
	userroleCmd.AddCommand(userroleExportUsersCmd)
	userroleImportUsersCmd.Flags().BoolP("delete-users", "D", true, "Delete users instead of delete roles for user.")
	userroleImportUsersCmd.Flags().BoolP("delete-homedir", "H", false, "Delete user home directory.  This option is used when the 'delete-users' option is enabled.")
	userroleImportUsersCmd.Flags().BoolP("update-password", "P", false, "Update user password.")
	userroleImportUsersCmd.Flags().StringP("default-password", "d", "", "Set the default password.  This option is used when the 'update-password' option is enabled.")
	userroleImportUsersCmd.Flags().IntP("header-size", "e", 0, "Set the header size.")
	userroleCmd.AddCommand(userroleImportUsersCmd)

	userrolerolesCmd.PersistentFlags().StringVarP(&roleTarget, "target", "t", "all", "Target roles.[all/standard/permission/system/extra]")
	userroleCmd.AddCommand(userrolerolesCmd)

	userroleUsersCmd.PersistentFlags().StringVarP(&userTarget, "target", "t", "all", "Target roles.[all/permission]")
	userroleCmd.AddCommand(userroleUsersCmd)

	userroleCmd.AddCommand(userroleAssignRoleToUserCmd)

	userroleCmd.AddCommand(userroleRemoveRoleFromUserCmd)
}
