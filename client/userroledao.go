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
