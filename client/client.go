package client

import (
	"fmt"

	"go.uber.org/zap"

	resty "gopkg.in/resty.v0"
)

// Logger is a logger for this package
var Logger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"./pentahotools.log"}
	config.Level.SetLevel(zap.DebugLevel)
	Logger, _ = config.Build()
}

// Client is the client class for pentaho.
type Client struct {
	url                  string
	User                 string
	Password             string
	client               *resty.Client
	JobClient            CarteClient
	TransformationClient CarteClient
}

// NewClient create new instance of Client.
func NewClient(url string, user string, password string) Client {
	Logger.Debug("NewClient", zap.String("url", url), zap.String("user", user), zap.String("password", "*****"))
	client := Client{
		url:      url,
		User:     user,
		Password: password,
	}
	client.client = resty.New().
		SetHostURL(url).
		SetBasicAuth(user, password).
		SetDisableWarn(true)
	client.JobClient = &JobClient{client.client}
	client.TransformationClient = &TransformationClient{client.client}
	return client
}

func (c Client) String() string {
	return fmt.Sprintf("Client(url=%s, user=%s)", c.url, c.User)
}
