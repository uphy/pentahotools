package cmd

import (
	"errors"
	"fmt"
	"strings"

	"strconv"

	set "github.com/deckarep/golang-set"
	pentahoclient "github.com/uphy/pentahotools/client"
)

// Permission represents a permission of the role
type Permission struct {
	name        string
	logicalName string
}

func (p Permission) String() string {
	return fmt.Sprintf("%s <%s>", p.name, p.logicalName)
}

// Role represents a role
type Role struct {
	name        string
	permissions *[]Permission
	immutable   bool
}

func (r Role) String() string {
	return fmt.Sprintf("Role (name=%s, immutable=%s, permissions=%s", r.name, strconv.FormatBool(r.immutable), *r.permissions)
}

// Print prints the role detail.
func (r *Role) Print() {
	immutable := ""
	if r.immutable {
		immutable = " (Immutable)"
	}
	fmt.Println(r.name + immutable)
	for _, permission := range *r.permissions {
		fmt.Printf("- %s\n", permission.String())
	}
}

// PermissionExecute is a permission to execute job/transformations.
var PermissionExecute = Permission{"execute", "org.pentaho.repository.execute"}

// PermissionDataSourceManagement is a permission to manage the datasources.
var PermissionDataSourceManagement = Permission{"datasource", "org.pentaho.platform.dataaccess.datasource.security.manage"}

// PermissionContentRead is a permission to read the contents.
var PermissionContentRead = Permission{"content-read", "org.pentaho.repository.read"}

// PermissionContentCreate is a permission to create the contents
var PermissionContentCreate = Permission{"content-create", "org.pentaho.repository.create"}

// PermissionContentSchedule is a permission to schedule the contents.
var PermissionContentSchedule = Permission{"schedule", "org.pentaho.scheduler.manage"}

// PermissionSecurityAdministration is a permission to administer the security.
var PermissionSecurityAdministration = Permission{"security", "org.pentaho.security.administerSecurity"}

// PermissionContentPublish is a permission to publish contents.
var PermissionContentPublish = Permission{"publish", "org.pentaho.security.publish"}

var allPermissions = []Permission{
	PermissionExecute,
	PermissionDataSourceManagement,
	PermissionContentRead,
	PermissionContentCreate,
	PermissionContentSchedule,
	PermissionSecurityAdministration,
	PermissionContentPublish,
}

type RoleClient struct {
	client *pentahoclient.Client
}

func NewRoleClient(client *pentahoclient.Client) *RoleClient {
	return &RoleClient{client}
}

func (c *RoleClient) getAvailablePermissionNames() []string {
	var names []string
	for _, p := range allPermissions {
		names = append(names, p.name)
	}
	return names
}

func (c *RoleClient) getPermissionFromLogicalName(logicalName string) *Permission {
	for _, permission := range allPermissions {
		if permission.logicalName == logicalName {
			return &permission
		}
	}
	return nil
}

func (c *RoleClient) getPermissionFromName(name string) *Permission {
	for _, permission := range allPermissions {
		if permission.name == name {
			return &permission
		}
	}
	return nil
}

func (c *RoleClient) newRole(assignment pentahoclient.Assignment) *Role {
	var permissions []Permission
	for _, logicalName := range assignment.LogicalRoles {
		permission := c.getPermissionFromLogicalName(logicalName)
		if permission == nil {
			c.client.Logger.Warn("no such permission: " + logicalName)
		}
		permissions = append(permissions, *c.getPermissionFromLogicalName(logicalName))
	}
	return &Role{
		name:        assignment.RoleName,
		immutable:   assignment.Immutable == "true",
		permissions: &permissions,
	}
}

// FindAllRoles finds all of the roles.
func (c *RoleClient) FindAllRoles() (*[]Role, error) {
	sytemRolesMap, err := c.client.ListPermissionsForRoles()
	if err != nil {
		return nil, err
	}
	var roles []Role
	for _, a := range sytemRolesMap.Assignments {
		roles = append(roles, *c.newRole(a))
	}
	return &roles, nil
}

// FindRole finds the specified role.
func (c *RoleClient) FindRole(role string) (*Role, error) {
	sytemRolesMap, err := c.client.ListPermissionsForRoles()
	if err != nil {
		return nil, err
	}
	for _, a := range sytemRolesMap.Assignments {
		if strings.ToLower(a.RoleName) == strings.ToLower(role) {
			return c.newRole(a), nil
		}
	}
	return nil, nil
}

// SetPermissionsOfRole set the permissions of the role.
func (c *RoleClient) SetPermissionsOfRole(role string, permissions ...string) error {
	var permissionNames []string
	for _, name := range permissions {
		p := c.getPermissionFromName(name)
		if p == nil {
			return c.unknownPermissionError(name)
		}
		permissionNames = append(permissionNames, p.logicalName)
	}
	return c.client.AssignPermissionsToRole(role, permissionNames...)
}

// AddPermissionsToRole add permissions to a role
func (c *RoleClient) AddPermissionsToRole(role string, permissions ...string) error {
	return c.addOrRemovePermissionsFromRole(true, role, permissions...)
}

// RemovePermissionsFromRole remove permissions to a role
func (c *RoleClient) RemovePermissionsFromRole(role string, permissions ...string) error {
	return c.addOrRemovePermissionsFromRole(false, role, permissions...)
}

func (c *RoleClient) addOrRemovePermissionsFromRole(add bool, role string, permissions ...string) error {
	r, err := c.FindRole(role)
	if err != nil {
		return err
	}
	if r == nil {
		return errors.New("no such role: " + role)
	}
	if r.immutable {
		return errors.New("immutable role: " + role)
	}
	s := set.NewSet()
	for _, p := range *r.permissions {
		s.Add(p.logicalName)
	}
	for _, p := range permissions {
		p2 := c.getPermissionFromName(p)
		if p2 == nil {
			return c.unknownPermissionError(p)
		}
		if add {
			if s.Contains(p2.logicalName) {
				return fmt.Errorf("'%s' already have the permission '%s'", role, p2.name)
			}
			s.Add(p2.logicalName)
		} else {
			if !s.Contains(p2.logicalName) {
				return fmt.Errorf("'%s' doesn't have the permission '%s'", role, p2.name)
			}
			s.Remove(p2.logicalName)
		}
	}
	var logicalPermissions []string
	for _, logicalPerission := range s.ToSlice() {
		logicalPermissions = append(logicalPermissions, logicalPerission.(string))
	}
	return c.client.AssignPermissionsToRole(role, logicalPermissions...)
}

func (c *RoleClient) unknownPermissionError(name string) error {
	return fmt.Errorf("unknown permission %s.  specify %s", name, c.getAvailablePermissionNames())
}
