package client

import (
	"errors"
	"fmt"
	"strings"

	"net/url"

	"go.uber.org/zap"
)

type user struct {
	UserName string `xml:"userName"`
	Password string `xml:"password"`
}

// CreateUser creates new pentaho user
func (c *Client) CreateUser(userName string, password string) error {
	c.Logger.Debug("CreateUser", zap.String("userName", userName), zap.String("password", "*****"))
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/xml").
		SetBody(user{userName, password}).
		Put("api/userroledao/createUser")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 400:
		return errors.New("Provided data has invalid format")
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 412:
		return errors.New("Unable to create user")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// CreateRole creates a role.
func (c *Client) CreateRole(roleName string) error {
	c.Logger.Debug("CreateRole", zap.String("roleName", roleName))
	resp, err := c.client.R().
		SetQueryParam("roleName", roleName).
		Put("api/userroledao/createRole?roleName=rName")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 400:
		return errors.New("Provided data has invalid format")
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 412:
		return errors.New("Unable to create role objects")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// DeleteRoles deletes roles
func (c *Client) DeleteRoles(roleNames ...string) error {
	c.Logger.Debug("DeleteRoles", zap.Strings("roleNames", roleNames))
	if len(roleNames) == 0 {
		return errors.New("Specify at least one role")
	}
	resp, err := c.client.R().
		SetQueryParam("roleNames", strings.Join(roleNames, "\t")+"\t").
		Put("api/userroledao/deleteRoles")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 500:
		return errors.New("The system was unable to delete the roles passed in")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// DeleteUsers deletes users
func (c *Client) DeleteUsers(userNames ...string) error {
	c.Logger.Debug("DeleteUsers", zap.Strings("userNames", userNames))
	if len(userNames) == 0 {
		return errors.New("Specify atleast one user")
	}
	resp, err := c.client.R().
		SetQueryParam("userNames", strings.Join(userNames, "\t")+"\t").
		Put("api/userroledao/deleteUsers")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 500:
		return errors.New("Internal server error prevented the system from properly retrieving either the user or roles")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ChangeUserPassword changes the password of the specified user.
func (c *Client) ChangeUserPassword(userName string, oldPassword string, newPassword string) error {
	c.Logger.Debug("ChangeUserPassword", zap.String("userName", userName), zap.String("oldPassword", "*****"), zap.String("newPassword", "*****"))
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"userName":    userName,
			"oldPassword": oldPassword,
			"newPassword": newPassword,
		}).
		Put("api/userroledao/user")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 400:
		return errors.New("Provided data has invalid format")
	case 403:
		return errors.New("Provided user name or password is incorrect")
	case 412:
		return errors.New("An error occurred in the platform")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// UpdatePassword changes the password of the specified user.
func (c *Client) UpdatePassword(userName string, password string) error {
	c.Logger.Debug("UpdatePassword", zap.String("userName", userName), zap.String("password", "*****"))
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"userName": userName,
			"password": url.QueryEscape(password),
		}).
		Put("api/userroledao/updatePassword")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 400:
		return errors.New("Provided data has invalid format")
	case 403:
		return errors.New("Provided user name or password is incorrect")
	case 412:
		return errors.New("An error occurred in the platform")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// AssignRolesToUser assigns a user to the specified roles.
func (c *Client) AssignRolesToUser(userName string, roles ...string) error {
	c.Logger.Debug("AssignRolesToUser", zap.String("userName", userName), zap.Strings("roles", roles))
	resp, err := c.client.R().
		SetQueryParam("userName", userName).
		SetQueryParam("roleNames", strings.Join(roles, "\t")).
		Put("api/userroledao/assignRoleToUser")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 500:
		return errors.New("Internal server error prevented the system from properly retrieving either the user or roles")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// RemoveRolesFromUser removes a user from the specified roles.
func (c *Client) RemoveRolesFromUser(userName string, roles ...string) error {
	c.Logger.Debug("RemoveRolesFromUser", zap.String("userName", userName), zap.Strings("roles", roles))
	resp, err := c.client.R().
		SetQueryParam("userName", userName).
		SetQueryParam("roleNames", strings.Join(roles, "\t")).
		Put("api/userroledao/removeRoleFromUser")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	case 500:
		return errors.New("Internal server error prevented the system from properly retrieving either the user or roles")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// ListPermissionsForRoles get the list of permissions for each roles.
func (c *Client) ListPermissionsForRoles() (*SystemRolesMap, error) {
	c.Logger.Debug("ListPermissionsForRoles")
	var result SystemRolesMap
	resp, err := c.client.R().
		SetQueryParam("locale", "en").
		SetResult(&result).
		SetHeader("Accept", "application/xml").
		Get("api/userroledao/logicalRoleMap")
	switch resp.StatusCode() {
	case 200:
		return &result, nil
	case 403:
		return nil, errors.New("Only users with administrative privileges can access this method")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// AssignPermissionsToRole assign permissions to the role.
func (c *Client) AssignPermissionsToRole(role string, permissions ...string) error {
	c.Logger.Debug("AssignPermissionsToRole")
	var m SystemRolesMap
	m.Assignments = append(m.Assignments, Assignment{
		RoleName:     role,
		LogicalRoles: permissions,
	})

	// add XMLName for modifying the tagname of the root xml element.
	body := struct {
		SystemRolesMap
		XMLName struct{} `xml:"systemRolesMap"`
	}{SystemRolesMap: m}

	resp, err := c.client.R().
		SetBody(body).
		SetHeader("Content-Type", "application/xml").
		Put("api/userroledao/roleAssignments")
	switch resp.StatusCode() {
	case 200:
		return nil
	case 403:
		return errors.New("Only users with administrative privileges can access this method")
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// Assignment represents a role information.
type Assignment struct {
	Immutable    string   `xml:"immutable"`
	LogicalRoles []string `xml:"logicalRoles"`
	RoleName     string   `xml:"roleName"`
}

// LocalizedRoleName reprensents the pair of the role name and localized role name.
type LocalizedRoleName struct {
	LocalizedName string `xml:"localizedName"`
	RoleName      string `xml:"roleName"`
}

// SystemRolesMap represents the permissions for the roles.
type SystemRolesMap struct {
	Assignments        []Assignment        `xml:"assignments"`
	LocalizedRoleNames []LocalizedRoleName `xml:"localizedRoleNames"`
}

func (m *SystemRolesMap) getLocalizedName(logicalRole string) string {
	for _, l := range m.LocalizedRoleNames {
		if l.RoleName == logicalRole {
			return l.LocalizedName
		}
	}
	return ""
}
