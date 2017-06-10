package cmd

import (
	"strings"

	"go.uber.org/zap"

	"fmt"

	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	client "github.com/uphy/pentahotools/client"
	"github.com/uphy/pentahotools/table"
	"gopkg.in/cheggaaa/pb.v1"
)

// DeleteUsersInFile delete users in file
func DeleteUsersInFile(file string, deleteHomeDirectory bool, bar *pb.ProgressBar) error {
	userTable, err := NewUserTable(file)
	if err != nil {
		return errors.Wrap(err, "reading file failed")
	}
	defer userTable.Close()
	bar.Prefix("Read file")
	users := []string{}
	for {
		userRow := userTable.Read()
		if userRow == nil {
			break
		}
		users = append(users, userRow.Name)
	}
	if deleteHomeDirectory {
		bar.Total = int64(1 + len(users))
	}
	return DeleteUsers(users, deleteHomeDirectory, bar)
}

// DeleteUsers deletes users
func DeleteUsers(users []string, deleteHomeDirectory bool, bar *pb.ProgressBar) error {
	if deleteHomeDirectory {
		bar.Total = int64(1 + len(users))
	} else {
		bar.Total = 1
	}
	bar.Prefix("Delete users")
	err := Client.DeleteUsers(users...)
	if err != nil {
		return errors.Wrap(err, "failed to delete users")
	}
	bar.Increment()
	if deleteHomeDirectory {
		for _, user := range users {
			homeDirectory := fmt.Sprintf("/home/%s", user)
			bar.Prefix("Delete home directory: " + homeDirectory)
			err = Client.DeleteFiles(homeDirectory)
			if err != nil {
				client.Logger.Warn("Failed to delete home directory.", zap.String("homeDirectory", homeDirectory))
			}
			bar.Increment()
		}
	}
	return nil
}

// CreateUsersInFile registers users from a file
func CreateUsersInFile(file string, bar *pb.ProgressBar) error {
	// getting user names from the repository
	allUsersSet, err := listExistingUsers()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of users")
	}
	// getting role names from the repository
	allRolesSet, err := listExistingRoles()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of roles")
	}

	userTable, err := NewUserTable(file)
	if err != nil {
		return errors.Wrap(err, "reading file failed")
	}
	defer userTable.Close()
	bar.Total = int64(userTable.GetCount())

	for {
		userRow := userTable.Read()
		if userRow == nil {
			break
		}
		bar.Prefix("User: " + userRow.Name)

		currentRoles, err := listRolesForUser(userRow.Name)
		if err != nil {
			client.Logger.Warn("Failed to list roles for user.", zap.String("user", userRow.Name), zap.Error(err))
			continue
		}

		// create user and change password if needed
		if allUsersSet.Contains(strings.ToLower(userRow.Name)) {
			if len(password) > 0 {
				err = Client.UpdatePassword(userRow.Name, userRow.Password)
				if err != nil {
					client.Logger.Warn("Failed to update password.", zap.String("user", userRow.Name), zap.Error(err))
				}
			}
		} else {
			var p = userRow.Password
			if len(p) == 0 {
				p = Client.Password
			}
			err = Client.CreateUser(userRow.Name, p)
			if err != nil {
				client.Logger.Warn("Failed to create user.", zap.String("user", userRow.Name), zap.Error(err))
			}
		}

		// assign roles
		assigningRoles := []string{}
		for _, role := range userRow.Roles {
			if !allRolesSet.Contains(role) {
				err = Client.CreateRole(role)
				if err != nil {
					client.Logger.Warn("Failed to create role.", zap.String("role", role), zap.Error(err))
				}
			}
			if !currentRoles.Contains(strings.ToLower(role)) {
				assigningRoles = append(assigningRoles, role)
			}
		}
		if len(assigningRoles) > 0 {
			err = Client.AssignRolesToUser(userRow.Name, assigningRoles...)
			if err != nil {
				client.Logger.Warn("Failed to assign role to user.", zap.String("user", userRow.Name), zap.Strings("roles", assigningRoles), zap.Error(err))
			}
		}
		// remove roles
		removingRoles := []string{}
		for _, role := range currentRoles.ToSlice() {
			roleString := role.(string)
			if !userRow.RoleSet.Contains(role) && strings.ToLower(roleString) != "authenticated" {
				removingRoles = append(removingRoles, roleString)
			}
		}
		if len(removingRoles) > 0 {
			err = Client.RemoveRolesFromUser(userRow.Name, removingRoles...)
			if err != nil {
				client.Logger.Warn("Failed to remove role from user.", zap.String("user", userRow.Name), zap.Strings("roles", assigningRoles), zap.Error(err))
			}
		}

		bar.Increment()
	}
	return nil
}

func listRolesForUser(userName string) (mapset.Set, error) {
	roles, err := Client.ListRolesForUser(userName)
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(roles), nil
}

func listExistingUsers() (mapset.Set, error) {
	users, err := Client.ListUsers()
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(users), nil
}

func listExistingRoles() (mapset.Set, error) {
	roles, err := Client.ListRoles()
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(roles), nil
}

func stringArrayToSetIgnoreCase(array *[]string) mapset.Set {
	set := mapset.NewSet()
	for _, elm := range *array {
		set.Add(strings.ToLower(elm))
	}
	return set
}

// NewUserTable read UserTable from a file.
func NewUserTable(file string) (*UserTable, error) {
	var row []string
	row = make([]string, 3) // 3 columns; username, role, password

	// scan table
	tmpTable, err := table.New(file)
	if err != nil {
		return nil, err
	}
	tmpUserTable := UserTable{table: tmpTable, row: row}
	defer tmpUserTable.Close()
	count := 0
	for {
		userRow := tmpUserTable.Read()
		if userRow == nil {
			break
		}
		count++
	}

	// create table
	table, err := table.New(file)
	if err != nil {
		return nil, err
	}
	return &UserTable{table, row, count}, nil
}

// UserTable represents a list of users from table structure files.
type UserTable struct {
	table table.Table
	row   []string
	count int
}

// GetCount gets the count of UserRows.
func (t *UserTable) GetCount() int {
	return t.count
}

// Read is a function reads a UserRow from the table.
func (t *UserTable) Read() *UserRow {
	for {
		if !t.table.ReadRow(&t.row) {
			return nil
		}
		userName := strings.TrimSpace(t.row[0])
		roles := strings.Split(strings.TrimSpace(t.row[1]), ":")
		roleSet := stringArrayToSetIgnoreCase(&roles)
		password := strings.TrimSpace(t.row[2])
		if len(userName) == 0 {
			continue
		}
		return &UserRow{userName, roleSet, roles, password}
	}
}

// Close closes the source file of the UserTable.
func (t *UserTable) Close() error {
	return t.table.Close()
}

// UserRow represents a row of the UserTable.
type UserRow struct {
	Name     string
	RoleSet  mapset.Set
	Roles    []string
	Password string
}
