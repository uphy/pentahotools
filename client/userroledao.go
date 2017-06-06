package pentahoclient

import (
	"errors"
	"fmt"
	"strings"
)

type user struct {
	UserName string `xml:"userName"`
	Password string `xml:"password"`
}

// CreateUser creates new pentaho user
func (c *Client) CreateUser(userName string, password string) error {
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

// DeleteUser deletes users
func (c *Client) DeleteUser(userNames ...string) error {
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

// AssignRoleToUser assigns a user to the specified roles.
func (c *Client) AssignRoleToUser(userName string, roles ...string) error {
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

// RemoveRoleFromUser removes a user from the specified roles.
func (c *Client) RemoveRoleFromUser(userName string, roles ...string) error {
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
