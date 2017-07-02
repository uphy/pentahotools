package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uphy/pentahotools/client"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "datasource",
		Short: "Datasource command.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.Aliases = []string{"ds"}
	RootCmd.AddCommand(cmd)

	/*
	 * Analysis Datasources
	 */
	cmd.AddCommand(&cobra.Command{
		Use:   "list-analysis-datasources",
		Short: "List the analysis datasources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := Client.ListAnalysisDatasources()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	})
	exportAnalysisDatasourceCmd := &cobra.Command{
		Use:   "export-analysis-datasource",
		Short: "Export an analysis datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var destination string
			switch len(args) {
			case 0:
				return errors.New("specify the datasource name and optionally destination path")
			case 1:
				destination = ""
			case 2:
				destination = args[1]
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			dest, err := Client.ExportAnalysisDatasource(args[0], destination, overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Saved the file as %s.\n", dest)
			return nil
		},
	}
	exportAnalysisDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(exportAnalysisDatasourceCmd)
	importAnalysisDatasourceCmd := &cobra.Command{
		Use:   "import-analysis-datasource",
		Short: "Import an analysis datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the source path")
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			return Client.ImportAnalysisDatasource(args[0], &client.ImportAnalysisDatasourceOptions{
				Overwrite: overwrite,
			})
		},
	}
	importAnalysisDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(importAnalysisDatasourceCmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "delete-analysis-datasource",
		Short: "Delete an analysis datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a datasource name")
			}
			return Client.DeleteAnalysisDatasource(args[0])
		},
	})

	/*
	 * JDBC Datasources
	 */
	cmd.AddCommand(&cobra.Command{
		Use:   "list-jdbc-datasources",
		Short: "List the JDBC datasources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := Client.ListJdbcDatasources()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	})
	exportJdbcDatasourceCmd := &cobra.Command{
		Use:   "export-jdbc-datasource",
		Short: "Export an jdbc datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var destination string
			switch len(args) {
			case 0:
				return errors.New("specify the datasource name and optionally destination path")
			case 1:
				destination = ""
			case 2:
				destination = args[1]
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			dest, err := Client.ExportJdbcDatasource(args[0], destination, overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Saved the file as %s.\n", dest)
			return nil
		},
	}
	exportJdbcDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(exportJdbcDatasourceCmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "import-jdbc-datasource",
		Short: "Import an JDBC datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the source path")
			}
			return Client.ImportJdbcDatasource(args[0])
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "delete-jdbc-datasource",
		Short: "Delete an JDBC datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a datasource name")
			}
			return Client.DeleteJdbcDatasource(args[0])
		},
	})
	/*
	 * Datasource Wizard Datasources
	 */
	cmd.AddCommand(&cobra.Command{
		Use:   "list-dsw-datasources",
		Short: "List the DSW datasources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := Client.ListDswDatasources()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	})
	exportDswDatasourceCmd := &cobra.Command{
		Use:   "export-dsw-datasource",
		Short: "Export an DSW datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var destination string
			switch len(args) {
			case 0:
				return errors.New("specify the datasource name and optionally destination path")
			case 1:
				destination = ""
			case 2:
				destination = args[1]
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			dest, err := Client.ExportDswDatasource(args[0], destination, overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Exported %s.\n", dest)
			return nil
		},
	}
	exportDswDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(exportDswDatasourceCmd)
	importDswDatasourceCmd := &cobra.Command{
		Use:   "import-dsw-datasource",
		Short: "Import an DSW datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the source path")
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			checkConnection, _ := cmd.Flags().GetBool("checkconnection")
			return Client.ImportDswDatasource(args[0], overwrite, checkConnection)
		},
	}
	importDswDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	importDswDatasourceCmd.Flags().BoolP("checkconnection", "c", false, "Check connection")
	cmd.AddCommand(importDswDatasourceCmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "delete-dsw-datasource",
		Short: "Delete an DSW datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a datasource name")
			}
			return Client.DeleteDswDatasource(args[0])
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "get-dsw-acl",
		Short: "Get the DSW ACL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a datasource name")
			}
			return Client.GetACLOfDswDatasource(args[0])
		},
	})
	/*
	 * Metadata Wizard Datasources
	 */
	cmd.AddCommand(&cobra.Command{
		Use:   "list-metadata-datasources",
		Short: "List the Metadata datasources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := Client.ListMetadataDatasources()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	})
	exportMetadataDatasourceCmd := &cobra.Command{
		Use:   "export-metadata-datasource",
		Short: "Export an Metadata datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var destination string
			switch len(args) {
			case 0:
				return errors.New("specify the datasource name and optionally destination path")
			case 1:
				destination = ""
			case 2:
				destination = args[1]
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			dest, err := Client.ExportMetadataDatasource(args[0], destination, overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Exported %s.\n", dest)
			return nil
		},
	}
	exportMetadataDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(exportMetadataDatasourceCmd)
	importMetadataDatasourceCmd := &cobra.Command{
		Use:   "import-metadata-datasource",
		Short: "Import an metadata datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the source path")
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			return Client.ImportMetadataDatasource(args[0], "", overwrite)
		},
	}
	importMetadataDatasourceCmd.Flags().BoolP("overwrite", "o", false, "overwrite if true")
	cmd.AddCommand(importMetadataDatasourceCmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "delete-metadata-datasource",
		Short: "Delete an Metadata datasource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a datasource name")
			}
			return Client.DeleteMetadataDatasource(args[0])
		},
	})
}
