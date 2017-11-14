package cmd

import (
	"errors"
	"os"

	"github.com/labstack/echo"
	"github.com/spf13/cobra"
)

func init() {
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Management console server",
		Long:  `Start the management console server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()
			staticRoot, err := findStaticRoot()
			if err != nil {
				return err
			}
			e.Static("/", staticRoot)
			e.Logger.Fatal(e.Start(":1323"))
			return nil
		},
	}
	RootCmd.AddCommand(serverCmd)
}

func findStaticRoot() (string, error) {
	if _, err := os.Stat("web/dist"); err == nil {
		return "web/dist", nil
	}
	if _, err := os.Stat("static"); err == nil {
		return "static", nil
	}
	return "", errors.New("static files not found")
}
