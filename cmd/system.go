package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// file
	var systemCmd = &cobra.Command{
		Use:   "system",
		Short: "System management command",
		Long:  `Manage system.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	RootCmd.AddCommand(systemCmd)

	systemCmd.AddCommand(&cobra.Command{
		Use:   "refresh-mondrian-schema-cache",
		Short: "Refresh the cache of mondrian.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Client.RefreshMondrianSchemaCache()
		},
	})
	systemCmd.AddCommand(&cobra.Command{
		Use:   "refresh-metadata",
		Short: "Refresh the metadata.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Client.RefreshMetadata()
		},
	})
	systemCmd.AddCommand(&cobra.Command{
		Use:   "refresh-reporting-data-cache",
		Short: "Refresh the reporting data cache.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Client.RefreshReportingDataCache()
		},
	})
	systemCmd.AddCommand(&cobra.Command{
		Use:   "refresh-system-settings",
		Short: "Refresh the system settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Client.RefreshSystemSettings()
		},
	})
}
