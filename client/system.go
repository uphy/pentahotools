package client

import (
	"fmt"
)

func (c *Client) get(path string) error {
	resp, err := c.client.R().Get(path)
	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("Unknown error. statusCode=%d", resp.StatusCode())
	}
}

// RefreshMondrianSchemaCache clears mondrian schema cache
func (c *Client) RefreshMondrianSchemaCache() error {
	c.Logger.Debug("RefreshMondrianSchemaCache")
	return c.get("api/system/refresh/mondrianSchemaCache")
}

// RefreshMetadata clears metadata cache
func (c *Client) RefreshMetadata() error {
	c.Logger.Debug("RefreshMetadata")
	return c.get("api/system/refresh/metadata")
}

// RefreshReportingDataCache clears reporting data cache
func (c *Client) RefreshReportingDataCache() error {
	c.Logger.Debug("RefreshReportingDataCache")
	return c.get("api/system/refresh/reportingDataCache")
}

// RefreshSystemSettings clears refresh system settings
func (c *Client) RefreshSystemSettings() error {
	c.Logger.Debug("RefreshSystemSettings")
	return c.get("api/system/refresh/systemSettings")
}
