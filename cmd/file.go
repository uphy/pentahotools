package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	// file
	var fileCmd = &cobra.Command{
		Use:   "file",
		Short: "File management command",
		Long:  `Manage files.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	RootCmd.AddCommand(fileCmd)

	// file tree
	var treeCmd = &cobra.Command{
		Use:   "tree",
		Short: "List the tree of files.",
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
			if err != nil {
				return err
			}
			root.Print()
			return nil
		},
	}
	treeCmd.Flags().BoolP("showHidden", "s", false, "Show hidden files")
	treeCmd.Flags().IntP("depth", "d", 1, "The depth of the tree")
	treeCmd.Aliases = []string{"ls"}
	fileCmd.AddCommand(treeCmd)

	// file backup
	fileCmd.AddCommand(&cobra.Command{
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
	})

	// file restore
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore Pentaho system.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify input file generated with 'backup' command")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			err := Client.Restore(args[0], overwrite)
			return err
		},
	}
	restoreCmd.Flags().BoolP("overwrite", "o", false, "If kept at the default of true, overwrites any value found on the system with the matching value that is being imported. ")
	fileCmd.AddCommand(restoreCmd)

	// file put
	fileCmd.AddCommand(&cobra.Command{
		Use:   "put",
		Short: "Put a file to the repository.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("specify input file and destination path")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Client.PutFile(args[0], args[1])
			return err
		},
	})
	// file get
	fileCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get a file from the repository.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("specify repository path and destination path")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Client.GetFile(args[0], args[1])
			return err
		},
	})

	// file delete
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete files from to the repository.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify the paths to delete")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Client.DeleteFiles(args...)
			return err
		},
	}
	deleteCmd.Flags().BoolP("permanent", "P", false, "Delete file permanently.")
	deleteCmd.Aliases = []string{"rm"}
	fileCmd.AddCommand(deleteCmd)

	// create-directory
	var createDirectoryCmd = &cobra.Command{
		Use:   "create-directory",
		Short: "Create a new directory.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the directory path")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Client.CreateDirectory(args[0])
			return err
		},
	}
	createDirectoryCmd.Aliases = []string{"mkdir"}
	fileCmd.AddCommand(createDirectoryCmd)
}
