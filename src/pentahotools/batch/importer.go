package batch

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/uphy/pentahotools/client"
)

func NewImporter(client BatchUserRoleClient, logger client.Logger) BatchUserRoleImporter {
	return BatchUserRoleImporter{
		client:              client,
		logger:              logger,
		appearedUserNames:   &map[string]string{},
		StrictCaseSensitive: true,
		CreateRoles:         false,
		UpdatePassword:      false,
	}
}

type BatchUserRoleImporter struct {
	client              BatchUserRoleClient
	allUserNames        *map[string]string
	allRoleNames        *map[string]string
	appearedUserNames   *map[string]string
	StrictCaseSensitive bool
	CreateRoles         bool
	UpdatePassword      bool
	logger              client.Logger
}

func (b *BatchUserRoleImporter) DeleteRemainingUsers(deleteHomeDirectory bool) error {
	var deletingUsers []string
	allUserNames, err := b.getAllUserNames()
	if err != nil {
		return errors.Wrap(err, "failed to list all user names.")
	}
	for k, v := range *allUserNames {
		if (*b.appearedUserNames)[k] == "" {
			if k == "admin" {
				b.logger.Warn("admin user can not be deleted.")
			} else {
				deletingUsers = append(deletingUsers, v)
			}
		}
	}
	if len(deletingUsers) == 0 {
		return nil
	}
	err = b.client.DeleteUsers(deletingUsers...)
	if err != nil {
		return err
	}
	if deleteHomeDirectory {
		errorCount := 0
		for _, user := range deletingUsers {
			folder := fmt.Sprintf("/home/%s", user)
			err := b.client.DeleteFiles(folder)
			if err != nil {
				b.logger.Error("Failed to delete the home directory:"+folder, zap.String("user", user), zap.Error(err))
				errorCount++
			}
		}
		if errorCount > 0 {
			return errors.New("failed to delete some home directory")
		}
	}
	return nil
}

func (b *BatchUserRoleImporter) Error(msg, user string, err error) error {
	b.logger.Error(msg, zap.String("user", user), zap.String("err", err.Error()))
	return err
}

func (b *BatchUserRoleImporter) Import(user string, password string, roles ...string) error {
	(*b.appearedUserNames)[strings.ToLower(user)] = user
	// Create user
	fixedUserName, err := b.getFixedUserNameOrCreate(user, password)
	if err != nil {
		return b.Error("Failed to get fixed username or create the user.", user, err)
	}
	if b.UpdatePassword {
		err := b.client.UpdatePassword(user, password)
		if err != nil {
			return b.Error("Failed to update password.", user, err)
		}
	}

	rolesForUser, err := b.client.ListRolesForUser(fixedUserName)
	if err != nil {
		return b.Error("Failed to list roles for user.", user, err)
	}

	// Assign roles
	assigningRoles := []string{}
assignLoop:
	for _, role := range roles {
		if len(role) == 0 {
			continue
		}
		fixedRoleName, err := b.getFixedRoleNameOrCreate(role)
		if err != nil {
			return b.Error("Failed to get the fixed role name or create the role.", user, err)
		}
		for _, r := range *rolesForUser {
			if fixedRoleName == r {
				// if the specified role have already been assigned, no need to assign.
				continue assignLoop
			}
		}
		assigningRoles = append(assigningRoles, fixedRoleName)
	}

	if len(assigningRoles) > 0 {
		err = b.client.AssignRolesToUser(fixedUserName, assigningRoles...)
		if err != nil {
			return b.Error("Failed to assign role to user.", fixedUserName, err)
		}
	}

	// Remove roles
	removingRoles := []string{}
removeLoop:
	for _, role := range *rolesForUser {
		if role == "Authenticated" {
			// do not delete authenticated
			continue
		}

		for _, r := range roles {
			fixedRoleName, err := b.getFixedRoleNameOrCreate(r)
			if err != nil {
				return err
			}
			if fixedRoleName == role {
				continue removeLoop
			}
		}
		removingRole := role
		if fixedUserName == "admin" && removingRole == "Administrator" {
			msg := "Can not remove 'Administrator' role from 'admin'."
			return b.Error(msg, fixedUserName, errors.New(msg))
		}
		removingRoles = append(removingRoles, removingRole)
	}
	if len(removingRoles) > 0 {
		err = b.client.RemoveRolesFromUser(fixedUserName, removingRoles...)
		if err != nil {
			return b.Error("Failed to remove role from user.", fixedUserName, err)
		}
	}
	return nil
}

func (b *BatchUserRoleImporter) getAllUserNames() (*map[string]string, error) {
	if b.allUserNames == nil {
		userNames, err := b.client.ListUsers()
		if err != nil {
			return nil, err
		}
		b.allUserNames = &map[string]string{}
		for _, userName := range *userNames {
			(*b.allUserNames)[strings.ToLower(userName)] = userName
		}
	}
	return b.allUserNames, nil
}

func (b *BatchUserRoleImporter) getFixedUserNameOrCreate(user string, password string) (string, error) {
	allUserNames, err := b.getAllUserNames()
	if err != nil {
		return "", err
	}
	userNameLower := strings.ToLower(user)
	fixedUserName := (*allUserNames)[userNameLower]
	if user != fixedUserName {
		if fixedUserName == "" {
			err := b.client.CreateUser(user, password)
			if err != nil {
				return "", err
			}
			fixedUserName = user
			(*allUserNames)[userNameLower] = user
		} else if b.StrictCaseSensitive {
			return "", fmt.Errorf("case of the user name mismatched.  got:%s, expected:%s", user, fixedUserName)
		}
	}
	return fixedUserName, nil
}

func (b *BatchUserRoleImporter) getAllRoleNames() (*map[string]string, error) {
	if b.allRoleNames == nil {
		roleNames, err := b.client.ListAllRoles()
		if err != nil {
			return nil, err
		}
		b.allRoleNames = &map[string]string{}
		for _, roleName := range *roleNames {
			(*b.allRoleNames)[strings.ToLower(roleName)] = roleName
		}
	}
	return b.allRoleNames, nil
}

func (b *BatchUserRoleImporter) getFixedRoleNameOrCreate(roleName string) (string, error) {
	allRoleNames, err := b.getAllRoleNames()
	if err != nil {
		return "", err
	}
	roleNameLower := strings.ToLower(roleName)
	fixedRoleName := (*allRoleNames)[roleNameLower]
	if roleName != fixedRoleName {
		if fixedRoleName == "" {
			if b.CreateRoles == false {
				return "", fmt.Errorf("No such role: %s", roleName)
			}
			err := b.client.CreateRole(roleName)
			if err != nil {
				return "", err
			}
			fixedRoleName = roleName
			(*allRoleNames)[roleNameLower] = roleName
		} else if b.StrictCaseSensitive && roleName != fixedRoleName {
			return "", fmt.Errorf("case of the role name mismatched.  got:%s, expected:%s", roleName, fixedRoleName)
		}
	}
	return fixedRoleName, nil
}
