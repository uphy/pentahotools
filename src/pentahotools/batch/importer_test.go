package batch

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock "github.com/uphy/pentahotools/mock_batch"
	mockclient "github.com/uphy/pentahotools/mock_client"
)

func newMockImporter(t *testing.T) (*mock.MockBatchUserRoleClient, *BatchUserRoleImporter, *mockclient.MockLogger) {
	ctrl := *gomock.NewController(t)
	logger := mockclient.NewMockLogger(&ctrl)
	c := mock.NewMockBatchUserRoleClient(&ctrl)
	importer := NewImporter(c, logger)
	return c, &importer, logger
}

func TestCreateNewUserAndRole(t *testing.T) {
	c, importer, _ := newMockImporter(t)
	importer.CreateRoles = true

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator"}, nil)
	c.EXPECT().CreateRole("roleA").Return(nil)
	c.EXPECT().AssignRolesToUser("user", "roleA").Return(nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateNewUserAndRoleError(t *testing.T) {
	c, importer, logger := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator"}, nil)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

	err := importer.Import("user", "password", "roleA")
	if err == nil {
		t.Log("Expected error")
		t.Fail()
	}
}

func TestCreateNewUserCaseMismatched(t *testing.T) {
	c, importer, logger := newMockImporter(t)

	importer.StrictCaseSensitive = true
	c.EXPECT().ListUsers().Return(&[]string{"user"}, nil)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
	err := importer.Import("User", "password", "roleA")
	if err == nil {
		t.Log("Importer is StrictCaseSensitiveMode but not return err.")
		t.Fail()
	}
}

func TestCreateNewUserAndExistingRole(t *testing.T) {
	c, importer, _ := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleA"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA"}, nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateNewUserAndExistingRoleCaseMismatched(t *testing.T) {
	c, importer, logger := newMockImporter(t)

	importer.StrictCaseSensitive = true
	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleA"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA"}, nil)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

	err := importer.Import("user", "password", "RoleA")
	if err == nil {
		t.Log("Importer is StrictCaseSensitiveMode but not return err.")
		t.Fail()
	}
}

func TestCreateNewUserAndAssignExistingRole(t *testing.T) {
	c, importer, _ := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA"}, nil)
	c.EXPECT().AssignRolesToUser("user", "roleA").Return(nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateNewUserAndRemoveRole(t *testing.T) {
	c, importer, _ := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleA", "roleB"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA", "roleB"}, nil)
	c.EXPECT().RemoveRolesFromUser("user", "roleB").Return(nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateNewUserAndRemoveRoleAdministrator(t *testing.T) {
	c, importer, logger := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("admin", "password").Return(nil)
	c.EXPECT().ListRolesForUser("admin").Return(&[]string{"roleA", "Administrator"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA"}, nil)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

	importer.Import("admin", "password", "roleA")
}

func TestCreateNewUserAndAssignRemoveRole(t *testing.T) {
	c, importer, _ := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleB"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA", "roleB"}, nil)
	c.EXPECT().AssignRolesToUser("user", "roleA").Return(nil)
	c.EXPECT().RemoveRolesFromUser("user", "roleB").Return(nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateExistingUser(t *testing.T) {
	c, importer, _ := newMockImporter(t)

	c.EXPECT().ListUsers().Return(&[]string{"user"}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleA"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA", "roleB"}, nil)

	importer.Import("user", "password", "roleA")
}

func TestCreateExistingUserUpdatePassword(t *testing.T) {
	c, importer, _ := newMockImporter(t)
	importer.UpdatePassword = true

	c.EXPECT().ListUsers().Return(&[]string{"user"}, nil)
	c.EXPECT().CreateUser("user", "password").Return(nil)
	c.EXPECT().UpdatePassword("user", "password")
	c.EXPECT().ListRolesForUser("user").Return(&[]string{"roleA"}, nil)
	c.EXPECT().ListAllRoles().Return(&[]string{"Administrator", "roleA", "roleB"}, nil)

	importer.Import("user", "password", "roleA")
}

func TestDeleteRemainingUsers(t *testing.T) {
	c, importer, logger := newMockImporter(t)
	importer.appearedUserNames = &map[string]string{
		"admin": "Admin",
		"user1": "User1",
		"user2": "User2",
	}
	c.EXPECT().ListUsers().Return(&[]string{"Admin", "User1", "User2", "User3"}, nil)
	c.EXPECT().DeleteUsers("User3").Return(nil)
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

	err := importer.DeleteRemainingUsers(false)
	if err != nil {
		t.Error("can't delete the user")
	}
}

func TestDeleteRemainingUsersAndHomeDirectories(t *testing.T) {
	c, importer, logger := newMockImporter(t)

	importer.appearedUserNames = &map[string]string{
		"admin": "Admin",
		"user1": "User1",
		"user2": "User2",
	}
	c.EXPECT().ListUsers().Return(&[]string{"Admin", "User1", "User2", "User3"}, nil)
	c.EXPECT().DeleteUsers("User3").Return(nil)
	c.EXPECT().DeleteFiles("/home/User3")
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

	err := importer.DeleteRemainingUsers(true)
	if err != nil {
		t.Error("can't delete the user")
	}
}

func TestDeleteRemainingUsersAdmin(t *testing.T) {
	c, importer, logger := newMockImporter(t)
	importer.appearedUserNames = &map[string]string{
		"user1": "User1",
		"user2": "User2",
	}
	c.EXPECT().ListUsers().Return(&[]string{"admin", "User1", "User2"}, nil)
	logger.EXPECT().Warn(gomock.Any())

	err := importer.DeleteRemainingUsers(false)
	if err != nil {
		t.Error("can't delete the user")
	}
}
