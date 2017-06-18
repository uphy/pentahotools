package client

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	resty "gopkg.in/resty.v0"
)

// Logger is a logger for this package
var Logger *CompositeLogger

func init() {
	fileConfig := zap.NewDevelopmentConfig()
	fileConfig.OutputPaths = []string{"./pentahotools.log"}
	fileConfig.Level.SetLevel(zap.DebugLevel)
	fileLogger, _ := fileConfig.Build()

	consoleConfig := zap.NewDevelopmentConfig()
	consoleConfig.Encoding = "console"
	consoleConfig.EncoderConfig.TimeKey = ""
	consoleConfig.EncoderConfig.StacktraceKey = ""
	consoleConfig.OutputPaths = []string{"stdout"}
	consoleConfig.Level.SetLevel(zap.WarnLevel)
	consoleLogger, _ := consoleConfig.Build()

	Logger = newCompositeLogger(*consoleLogger, *fileLogger)
}

// CompositeLogger merges multiple loggers.
type CompositeLogger struct {
	Logger []zap.Logger
}

func newCompositeLogger(logger ...zap.Logger) *CompositeLogger {
	return &CompositeLogger{logger}
}

// Debug records the debug logs
func (l *CompositeLogger) Debug(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Debug(msg, fields...)
	}
}

// Warn records the debug logs
func (l *CompositeLogger) Warn(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Warn(msg, fields...)
	}
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
