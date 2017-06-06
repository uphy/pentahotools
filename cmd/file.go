package cmd

import (
	"fmt"

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
		root, err := Client.Tree(path, 1, true)
		fmt.Println(root.File.CreatedDate)
		return err
	},
}

func init() {
	RootCmd.AddCommand(fileCmd)
	fileCmd.AddCommand(fileLsCmd)
}
