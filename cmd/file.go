package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File management command",
	Long:  `Manage files.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var fileLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List files.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		if len(args) > 0 {
			path = args[0]
		} else {
			path = "/"
		}
		showHidden, _ := cmd.Flags().GetBool("showHidden")
		depth, _ := cmd.Flags().GetInt("depth")
		root, err := Client.Tree(path, depth, showHidden)
		root.Print()
		return err
	},
}

var fileBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Pentaho system.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("specify output zip path")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Client.Backup(args[0])
		return err
	},
}

var overwrite bool
var fileRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore Pentaho system.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("specify input file generated with 'backup' command")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Client.Restore(args[0], overwrite)
		return err
	},
}

func init() {
	RootCmd.AddCommand(fileCmd)
	fileLsCmd.Flags().BoolP("showHidden", "s", false, "Show hidden files")
	fileLsCmd.Flags().IntP("depth", "d", 1, "The depth of the tree")
	fileCmd.AddCommand(fileLsCmd)
	fileCmd.AddCommand(fileBackupCmd)

	fileRestoreCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "If kept at the default of true, overwrites any value found on the system with the matching value that is being imported. ")
	fileCmd.AddCommand(fileRestoreCmd)
}
