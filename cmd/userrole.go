package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	mapset "github.com/deckarep/golang-set"
	"github.com/spf13/cobra"
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

var file string

var userroleCreateUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create user",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if file != "" {
			if len(args) > 0 {
				return errors.New("can not set both option(file) and arguments")
			}
		} else if len(args) != 2 {
			return errors.New("Specify 2 arguments(username and password)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if file == "" {
			return Client.CreateUser(args[0], args[1])
		}
		// Create users with CSV file.
		bar := pb.StartNew(0)
		bar.Prefix("Listing existing users...")
		existingUsers, err := listExistingUsers()
		if err != nil {
			bar.FinishPrint("Failed to list existing users.")
			return errors.Wrap(err, "Failed to list existing users.")
		}

		bar.Prefix("Scanning CSV file...")
		users, err := ReadUsersFile(file)
		if err != nil {
			return errors.Wrap(err, "failed to read csv file.")
		}

		bar.Total = int64(len(users))
		bar.Start()
		skipped := mapset.NewSet()
		for _, u := range users {
			if existingUsers.Contains(u.name) {
				bar.Prefix("Already exist: " + u.name)
				bar.Increment()
				skipped.Add(u.name)
				continue
			}
			bar.Prefix("Create user: " + u.name)
			err = Client.CreateUser(u.name, u.password)
			if err != nil {
				bar.FinishPrint("Failed to create user: " + u.name)
				return err
			}
		}
		bar.FinishPrint("Finished to create the users.")
		if skipped.Cardinality() > 0 {
			fmt.Printf("Skipped: %s\n", skipped.ToSlice())
		}
		return nil
	},
}

func listExistingUsers() (mapset.Set, error) {
	existingUsers := mapset.NewSet()
	users, err := Client.ListUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range *users {
		existingUsers.Add(u)
	}
	return existingUsers, nil
}

var userroleDeleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete the specified users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if file == "" {
			return Client.DeleteUser(args...)
		}

		bar := pb.StartNew(0)
		bar.Prefix("Listing existing users...")
		existingUsers, err := listExistingUsers()
		if err != nil {
			bar.FinishPrint("Failed to list existing users.")
			return errors.Wrap(err, "Failed to list existing users.")
		}

		bar.Prefix("Scanning CSV file...")
		users, err := ReadUsersFile(file)
		if err != nil {
			return errors.Wrap(err, "failed to read csv file.")
		}
		bar.Total = int64(len(users))
		bar.Start()
		skipped := mapset.NewSet()
		for _, u := range users {
			if !existingUsers.Contains(u.name) {
				bar.Prefix("Not exist: " + u.name)
				bar.Increment()
				skipped.Add(u.name)
				continue
			}
			bar.Prefix("Delete user: " + u.name)
			err = Client.DeleteUser(u.name, u.password)
			if err != nil {
				bar.FinishPrint("Failed to delete user: " + u.name)
				return err
			}
		}
		bar.FinishPrint("Finished to delete the users.")
		if skipped.Cardinality() > 0 {
			fmt.Printf("Skipped: %s\n", skipped.ToSlice())
		}
		return nil
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
		return Client.AssignRoleToUser(userName, roles...)
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
		return Client.RemoveRoleFromUser(userName, roles...)
	},
}

func init() {
	RootCmd.AddCommand(userroleCmd)

	userroleCmd.AddCommand(userroleChangeUserPasswordCmd)

	userroleCreateUserCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "Batch create from CSV file.")
	userroleCmd.AddCommand(userroleCreateUserCmd)

	userroleDeleteUserCmd.Flags().StringVarP(&file, "file", "f", "", "Batch delete from CSV file.")
	userroleCmd.AddCommand(userroleDeleteUserCmd)

	userrolerolesCmd.PersistentFlags().StringVarP(&roleTarget, "target", "t", "all", "Target roles.[all/standard/permission/system/extra]")
	userroleCmd.AddCommand(userrolerolesCmd)

	userroleUsersCmd.PersistentFlags().StringVarP(&userTarget, "target", "t", "all", "Target roles.[all/permission]")
	userroleCmd.AddCommand(userroleUsersCmd)

	userroleCmd.AddCommand(userroleAssignRoleToUserCmd)

	userroleCmd.AddCommand(userroleRemoveRoleFromUserCmd)
}
