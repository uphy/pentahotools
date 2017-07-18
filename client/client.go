package client

import (
	"fmt"

	"go.uber.org/zap"

	resty "gopkg.in/resty.v0"
)

// Client is the client class for pentaho.
type Client struct {
	url                  string
	User                 string
	Password             string
	client               *resty.Client
	JobClient            CarteClient
	TransformationClient CarteClient
	Logger               Logger
}

// NewClient create new instance of Client.
func NewClient(url string, user string, password string) Client {
	logger := NewCompositeLogger()
	logger.Debug("NewClient", zap.String("url", url), zap.String("user", user), zap.String("password", "*****"))
	client := Client{
		url:      url,
		User:     user,
		Password: password,
		Logger:   logger,
	}
	client.client = resty.New().
		SetHostURL(url).
		SetBasicAuth(user, password).
		SetDisableWarn(true)
	client.JobClient = &JobClient{client.client, logger}
	client.TransformationClient = &TransformationClient{client.client, logger}
	return client
}

func (c Client) String() string {
	return fmt.Sprintf("Client(url=%s, user=%s)", c.url, c.User)
}
