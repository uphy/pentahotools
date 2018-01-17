package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	SlaveServer struct {
		Name          string `xml:"name"`
		HostName      string `xml:"hostname"`
		Port          int    `xml:"port"`
		WebAppName    string `xml:"webappname"`
		UserName      string `xml:"username"`
		Password      string `xml:"password"`
		ProxyHostName string `xml:"proxy_hostname"`
		ProxyPort     string `xml:"proxy_port"`
		NonProxyHosts string `xml:"non_proxy_hosts"`
		Master        string `xml:"master"`
	}
	SlaveServerDetection struct {
		SlaveServer      SlaveServer `xml:"slaveserver"`
		Active           string      `xml:"active"`
		LastActiveDate   string      `xml:"last_active_date"`
		LastInactiveDate string      `xml:"last_inactive_date"`
	}
	SlaveServerDetections struct {
		SlaveServerDetections []SlaveServerDetection `xml:"slaveserverdetection"`
	}
)

func (c *Client) GetSlaves() (*SlaveServerDetections, error) {
	c.Logger.Debug("GetSlaves")
	var slaveServerDetections SlaveServerDetections
	resp, err := c.client.R().
		SetResult(&slaveServerDetections).
		Get(fmt.Sprintf("kettle/getSlaves"))
	switch resp.StatusCode() {
	case 200:
		return &slaveServerDetections, nil
	case 500:
		return nil, errors.New("Internal server error occurs during request processing")
	default:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}
