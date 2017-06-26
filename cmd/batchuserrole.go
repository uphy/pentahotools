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

// ExportUsers exports user list to the file.
func ExportUsers(file string, withHeader bool, bar *pb.ProgressBar) error {
	bar.Prefix("List users")
	users, err := Client.ListUsers()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of users.")
	}
	bar.Total = int64(len(*users))

	writer, err := table.NewWriter(file)
	if err != nil {
		return err
	}
	defer writer.Close()
	if withHeader {
		writer.WriteHeader(&[]string{"User", "Roles"})
	}
	for _, user := range *users {
		bar.Prefix("Roles for " + user)
		roles, err := Client.ListRolesForUser(user)
		if err != nil {
			client.Logger.Warn("Failed to list roles for user.", zap.String("user", user))
			bar.Increment()
			continue
		}
		var filteredRoles []string
		for _, role := range *roles {
			if role == "Authenticated" {
				continue
			}
			filteredRoles = append(filteredRoles, role)
		}
		writer.WriteRow(&[]string{user, strings.Join(filteredRoles, ":")})
		bar.Increment()
	}
	return nil
}

// DeleteUsersInFile delete users in file
func DeleteUsersInFile(file string, deleteHomeDirectory bool, headers int, bar *pb.ProgressBar) error {
	userTable, err := NewUserTable(file, headers)
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

// ImportUsers registers users from a file
func ImportUsers(file string, options *ImportUsersOptions, bar *pb.ProgressBar) error {
	// getting user names from the repository
	allUsersSetLower, err := listExistingUsers()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of users")
	}
	usersNotInFileLower := allUsersSetLower.Clone()
	// getting role names from the repository
	allRolesSetLower, err := listExistingRoles()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of roles")
	}

	userTable, err := NewUserTable(file, options.HeaderSize)
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
		usersNotInFileLower.Remove(strings.ToLower(userRow.Name))

		var currentRolesLower mapset.Set
		var currentRolesMap map[string]string
		if allUsersSetLower.Contains(strings.ToLower(userRow.Name)) {
			currentRolesLower, currentRolesMap, err = listRolesForUser(userRow.Name)
			if err != nil {
				client.Logger.Warn("Failed to list roles for user.", zap.String("user", userRow.Name), zap.Error(err))
				continue
			}
		} else {
			currentRolesLower = mapset.NewSet()
			currentRolesMap = map[string]string{}
		}

		// create user and change password if needed
		password := userRow.Password
		if len(password) == 0 {
			if len(options.DefaultPassword) > 0 {
				password = options.DefaultPassword
			} else {
				password = Client.Password
			}
		}
		if allUsersSetLower.Contains(strings.ToLower(userRow.Name)) {
			if options.UpdatePassword {
				err = Client.UpdatePassword(userRow.Name, password)
				if err != nil {
					client.Logger.Warn("Failed to update password.", zap.String("user", userRow.Name), zap.Error(err))
				}
			}
		} else {
			err = Client.CreateUser(userRow.Name, password)
			if err != nil {
				client.Logger.Warn("Failed to create user.", zap.String("user", userRow.Name), zap.Error(err))
			}
		}

		// assign roles
		assigningRoles := []string{}
		for _, role := range userRow.Roles {
			if len(role) == 0 {
				continue
			}
			roleLower := strings.ToLower(role)
			if !allRolesSetLower.Contains(roleLower) {
				err = Client.CreateRole(role)
				if err != nil {
					client.Logger.Warn("Failed to create role.", zap.String("role", role), zap.Error(err))
				}
			}
			if !currentRolesLower.Contains(roleLower) {
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
		for _, roleLower := range currentRolesLower.ToSlice() {
			roleLowerString := roleLower.(string)
			if !userRow.RoleSet.Contains(roleLower) && roleLowerString != "authenticated" {
				roleOriginal := currentRolesMap[roleLowerString]
				if strings.ToLower(userRow.Name) == "admin" && roleLower == "administrator" {
					client.Logger.Warn("User 'admin' should be 'administrator'.")
					continue
				}
				removingRoles = append(removingRoles, roleOriginal)
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
	// Delete users not in the input file
	usersNotInFileLower.Remove("admin")
	if usersNotInFileLower.Cardinality() > 0 {
		if options.DeleteUsers {
			deletingUsers := make([]string, usersNotInFileLower.Cardinality())
			for i, u := range usersNotInFileLower.ToSlice() {
				deletingUsers[i] = u.(string)
			}
			DeleteUsers(deletingUsers, options.DeleteHomeDirectory, bar)
		} else {
			bar.Total += int64(usersNotInFileLower.Cardinality())
			for _, user := range usersNotInFileLower.ToSlice() {
				userString := user.(string)

				bar.Prefix("Remove roles: " + userString)
				rolesSetLower, roleMap, err := listRolesForUser(userString)
				if err != nil {
					client.Logger.Warn("Failed to list the roles of the user which doesn't exist in the input file.", zap.String("user", userString), zap.String("file", file))
					bar.Increment()
					continue
				}
				var filteredRoles []string
				for _, role := range rolesSetLower.ToSlice() {
					if role == "authenticated" {
						continue
					}
					filteredRoles = append(filteredRoles, roleMap[role.(string)])
				}
				err = Client.RemoveRolesFromUser(userString, filteredRoles...)
				if err != nil {
					client.Logger.Warn("Failed to remove roles of the user which doesn't exist in the input file.", zap.String("user", userString), zap.String("roles", rolesSetLower.String()), zap.String("file", file))
					bar.Increment()
					continue
				}
				bar.Increment()
			}
		}
	}
	return nil
}

// ImportUsersOptions represents the options for ImportUsers func.
type ImportUsersOptions struct {
	DeleteUsers         bool
	DeleteHomeDirectory bool
	UpdatePassword      bool
	DefaultPassword     string
	HeaderSize          int
}

func listRolesForUser(userName string) (mapset.Set, map[string]string, error) {
	roles, err := Client.ListRolesForUser(userName)
	if err != nil {
		return nil, nil, err
	}
	lowerToOriginalMap := map[string]string{}
	return stringArrayToSetIgnoreCase(roles, lowerToOriginalMap), lowerToOriginalMap, nil
}

func listExistingUsers() (mapset.Set, error) {
	users, err := Client.ListUsers()
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(users, nil), nil
}

func listExistingRoles() (mapset.Set, error) {
	roles, err := Client.ListAllRoles()
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(roles, nil), nil
}

func stringArrayToSetIgnoreCase(array *[]string, lowerToOriginalMap map[string]string) mapset.Set {
	set := mapset.NewSet()
	for _, elm := range *array {
		lower := strings.ToLower(elm)
		set.Add(lower)
		if lowerToOriginalMap != nil {
			lowerToOriginalMap[lower] = elm
		}

	}
	return set
}

// NewUserTable read UserTable from a file.
func NewUserTable(file string, headers int) (*UserTable, error) {
	var row []string
	row = make([]string, 3) // 3 columns; username, role, password

	// scan table
	tmpTable, err := table.NewReader(file)
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
	table, err := table.NewReader(file)
	if err != nil {
		return nil, err
	}
	// skip headers
	for i := 0; i < headers; i++ {
		table.ReadRow(&row)
	}
	return &UserTable{table, row, count}, nil
}

// UserTable represents a list of users from table structure files.
type UserTable struct {
	table table.Reader
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
		roleSet := stringArrayToSetIgnoreCase(&roles, nil)
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
