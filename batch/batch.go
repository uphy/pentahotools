package batch

import (
	"strings"

	"go.uber.org/zap"

	"fmt"

	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"github.com/uphy/pentahotools/client"
	"github.com/uphy/pentahotools/table"
	"gopkg.in/cheggaaa/pb.v1"
)

// ExportUsers exports user list to the file.
func ExportUsers(file string, options *ExportUsersOptions, bar *pb.ProgressBar, bclient BatchUserRoleClient, logger client.Logger) error {
	bar.Prefix("List users")
	users, err := bclient.ListUsers()
	if err != nil {
		return errors.Wrap(err, "failed to get the list of users.")
	}
	bar.Total = int64(len(*users))

	writerOptions := map[int]string{}
	writerOptions[table.CsvSeparator] = options.Separator
	writer, err := table.NewWriter(file, writerOptions)
	if err != nil {
		return err
	}
	defer writer.Close()
	if options.WithHeader {
		writer.WriteHeader(&[]string{"User", "Roles"})
	}
	for _, user := range *users {
		bar.Prefix("Roles for " + user)
		roles, err := bclient.ListRolesForUser(user)
		if err != nil {
			logger.Warn("Failed to list roles for user.", zap.String("user", user))
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

// DeleteUsers deletes users
func DeleteUsers(users []string, deleteHomeDirectory bool, bar *pb.ProgressBar, bclient BatchUserRoleClient, logger client.Logger) error {
	if deleteHomeDirectory {
		bar.Total = int64(1 + len(users))
	} else {
		bar.Total = 1
	}
	bar.Prefix("Delete users")
	err := bclient.DeleteUsers(users...)
	if err != nil {
		return errors.Wrap(err, "failed to delete users")
	}
	bar.Increment()
	if deleteHomeDirectory {
		for _, user := range users {
			homeDirectory := fmt.Sprintf("/home/%s", user)
			bar.Prefix("Delete home directory: " + homeDirectory)
			err = bclient.DeleteFiles(homeDirectory)
			if err != nil {
				logger.Warn("Failed to delete home directory.", zap.String("homeDirectory", homeDirectory))
			}
			bar.Increment()
		}
	}
	return nil
}

// ImportUsers registers users from a file
func ImportUsers(file string, options *ImportUsersOptions, bar *pb.ProgressBar, bclient BatchUserRoleClient, logger client.Logger) error {
	// validation
	clientForValidation := NewBatchUserRoleClientForValidation(bclient)
	err := importUsers(file, options, bar, clientForValidation, logger)
	if err != nil {
		return errors.Wrap(err, "Validation failure")
	}
	bar.Set(0)
	// import
	return importUsers(file, options, bar, bclient, logger)
}

// ImportUsers registers users from a file
func importUsers(file string, options *ImportUsersOptions, bar *pb.ProgressBar, bclient BatchUserRoleClient, logger client.Logger) error {
	importer := NewImporter(bclient, logger)
	importer.CreateRoles = options.CreateRoles
	importer.UpdatePassword = options.UpdatePassword

	userTableOptions := map[int]string{}
	userTableOptions[table.CommonHeaderSize] = fmt.Sprint(options.HeaderSize)
	userTableOptions[table.CsvSeparator] = fmt.Sprint(options.Separator)
	userTable, err := NewUserTable(file, userTableOptions)
	if err != nil {
		return errors.Wrap(err, "reading file failed")
	}
	defer userTable.Close()
	bar.Total = int64(userTable.GetCount())

	var hasErrorInImport bool
	for {
		userRow := userTable.Read()
		if userRow == nil {
			break
		}
		bar.Prefix("User: " + userRow.Name)

		// create user and change password if needed
		password := userRow.Password
		if len(password) == 0 {
			password = options.DefaultPassword
		}
		err := importer.Import(userRow.Name, password, userRow.Roles...)
		if err != nil {
			hasErrorInImport = true
			logger.Error("Failed to import the user.", zap.String("user", userRow.Name), zap.Strings("roles", userRow.Roles), zap.String("err", err.Error()))
		}
		bar.Increment()
	}
	if hasErrorInImport {
		return errors.New("Errors occured while importing users")
	}
	if options.DeleteUsers == false {
		return nil
	}
	return importer.DeleteRemainingUsers(options.DeleteHomeDirectory)
}

type ExportUsersOptions struct {
	Separator  string
	WithHeader bool
}

// ImportUsersOptions represents the options for ImportUsers func.
type ImportUsersOptions struct {
	DeleteUsers         bool
	DeleteHomeDirectory bool
	UpdatePassword      bool
	DefaultPassword     string
	HeaderSize          int
	Separator           string
	CreateRoles         bool
}

func listRolesForUser(bclient BatchUserRoleClient, userName string) (mapset.Set, map[string]string, error) {
	roles, err := bclient.ListRolesForUser(userName)
	if err != nil {
		return nil, nil, err
	}
	lowerToOriginalMap := map[string]string{}
	return stringArrayToSetIgnoreCase(roles, lowerToOriginalMap), lowerToOriginalMap, nil
}

func listExistingUsers(bclient BatchUserRoleClient) (mapset.Set, error) {
	users, err := bclient.ListUsers()
	if err != nil {
		return nil, err
	}
	return stringArrayToSetIgnoreCase(users, nil), nil
}

func listExistingRoles(bclient BatchUserRoleClient) (mapset.Set, error) {
	roles, err := bclient.ListAllRoles()
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
func NewUserTable(file string, options map[int]string) (*UserTable, error) {
	var row []string
	row = make([]string, 3) // 3 columns; username, role, password

	// scan table
	tmpTable, err := table.NewReader(file, nil)
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
	table, err := table.NewReader(file, options)
	if err != nil {
		return nil, err
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
