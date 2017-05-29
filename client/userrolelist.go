package pentahoclient

type userList struct {
	Users []string `xml:"users"`
}

type users struct {
	Users []string `xml:"user"`
}

type roleList struct {
	Roles []string `xml:"roles"`
}

type roles struct {
	Roles []string `xml:"role"`
}

// ListUsers function lists all users.
func (c *Client) ListUsers() (*[]string, error) {
	logger.Info("ListUser")
	var t userList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/users")
	return &t.Users, err
}

// ListPermissionUsers function lists all users.
func (c *Client) ListPermissionUsers() (*[]string, error) {
	var t userList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/permission-users")
	return &t.Users, err
}

// ListUsersInRole lists the users in role
func (c *Client) ListUsersInRole(role string) (*[]string, error) {
	var t users
	_, err := c.client.R().
		SetResult(&t).
		SetQueryParam("role", role).
		Get("api/userrolelist/getUsersInRole")
	return &t.Users, err
}

// ListRoles function lists all roles.
func (c *Client) ListRoles() (*[]string, error) {
	var t roleList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/roles")
	return &t.Roles, err
}

// ListAllRoles function lists all roles.
func (c *Client) ListAllRoles() (*[]string, error) {
	var t roleList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/allRoles")
	return &t.Roles, err
}

// ListExtraRoles function lists all roles.
func (c *Client) ListExtraRoles() (*[]string, error) {
	var t roleList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/extraRoles")
	return &t.Roles, err
}

// ListPermissionRoles function lists all roles.
func (c *Client) ListPermissionRoles() (*[]string, error) {
	var t roleList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/permission-roles")
	return &t.Roles, err
}

// ListSystemRoles function lists all roles.
func (c *Client) ListSystemRoles() (*[]string, error) {
	var t roleList
	_, err := c.client.R().
		SetResult(&t).
		Get("api/userrolelist/systemRoles")
	return &t.Roles, err
}

// ListRolesForUser function lists all roles for the specified user.
func (c *Client) ListRolesForUser(user string) (*[]string, error) {
	var t roles
	_, err := c.client.R().
		SetResult(&t).
		SetQueryParam("user", user).
		Get("api/userrolelist/getRolesForUser")
	return &t.Roles, err
}
