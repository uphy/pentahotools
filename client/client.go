package pentahoclient

import (
	"fmt"

	resty "gopkg.in/resty.v0"
)

// Client is the client class for pentaho.
type Client struct {
	url    string
	user   string
	client *resty.Client
}

// NewClient create new instance of Client.
func NewClient(url string, user string, password string) Client {
	client := Client{
		url:  url,
		user: user,
	}
	client.client = resty.New().
		SetHostURL(url).
		SetBasicAuth(user, password).
		SetDisableWarn(true)
	return client
}

func (c Client) String() string {
	return fmt.Sprintf("Client(url=%s, user=%s)", c.url, c.user)
}
