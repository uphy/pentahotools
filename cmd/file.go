package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/uphy/pentahotools/client"
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
	// file download
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download the content of the file/folder in the repository.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a repository path and optinally destination path")
			}
			if len(args) > 2 {
				return errors.New("too many arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repositoryFile := args[0]
			var destination string
			if len(args) > 1 {
				destination = args[1]
			} else {
				destination = ""
			}
			withManifest, _ := cmd.Flags().GetBool("manifest")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			path, err := Client.DownloadFile(repositoryFile, destination, withManifest, overwrite)
			if err != nil {
				return err
			}
			fmt.Println("Saved file to " + path)
			return nil
		},
	}
	downloadCmd.Aliases = []string{"dl"}
	downloadCmd.Flags().BoolP("overwrite", "o", false, "overwrite if exist")
	downloadCmd.Flags().BoolP("manifest", "m", false, "with manifest")
	fileCmd.AddCommand(downloadCmd)

	// file import
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Attempts to import all files from the zip archive or single file.  A log file is produced at the end of import.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("specify a upload file path and destination repository path")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, filename := filepath.Split(args[0])
			err := Client.ImportFile(args[0], args[1], &client.ImportParameters{
				OverwriteFile:           true,
				LogLevel:                "ERROR",
				FileNameOverride:        filename,
				RetainOwnership:         true,
				OverwriteACLPermissions: false,
				ApplyACLPermissions:     false,
				Charset:                 "",
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
	//importCmd.Flags().BoolP("overwrite", "o", false, "overwrite if exist")
	//importCmd.Flags().StringP("loglevel", "L", string(client.LogLevels.Basic), "log level")
	fileCmd.AddCommand(importCmd)

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

	// clear-cache
	fileCmd.AddCommand(&cobra.Command{
		Use:   "clear-cache",
		Short: "Clear the cache of Analyzer and Mondrian.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a catalog name")
			}
			return Client.ClearCache(args[0])
		},
	})

	// set-owner
	fileCmd.AddCommand(&cobra.Command{
		Use:   "set-owner",
		Short: "Set the owner of file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("specify the resource path and owner user name")
			}
			path := args[0]
			owner := args[1]
			acl, err := Client.GetACL(path)
			if err != nil {
				return err
			}
			acl.Owner = owner
			return Client.SetACL(path, acl)
		},
	})

	// get-acl
	fileCmd.AddCommand(&cobra.Command{
		Use:   "get-acl",
		Short: "Get the access control list of file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the resource path")
			}
			path := args[0]
			acl, err := Client.GetACL(path)
			if err != nil {
				return err
			}
			fmt.Printf("ID                : %s\n", acl.ID)
			fmt.Printf("Entries Inheriting: %s\n", acl.EntriesInheriting)
			fmt.Printf("Owner             : %s\n", acl.Owner)
			fmt.Printf("Owner type        : %s\n", acl.OwnerType)
			for _, ac := range acl.Aces {
				fmt.Printf("[%s]\n", ac.Recipient)
				fmt.Printf("Recipient Type: %s\n", ac.RecipientType)
				fmt.Printf("Permissions   : %s\n", ac.PermissionsString())
				fmt.Printf("Modifiable    : %s\n", ac.Modifiable)
			}
			return nil
		},
	})

	// set-acl
	setACLCmd := &cobra.Command{
		Use:   "set-acl",
		Short: "Set the access control list of file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the resource path")
			}
			path := args[0]
			acl, err := Client.GetACL(path)
			if err != nil {
				return err
			}
			if cmd.Flag("inherit").Changed {
				inherit, _ := cmd.Flags().GetBool("inherit")
				acl.EntriesInheriting = fmt.Sprint(inherit)
				if inherit {
					acl.Aces = nil
				}
			}
			if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
				acl.Owner = owner
			}
			if ownerType, _ := cmd.Flags().GetString("ownertype"); ownerType != "" {
				acl.OwnerType = ownerType
			}
			return Client.SetACL(path, acl)
		},
	}
	setACLCmd.Flags().BoolP("inherit", "i", true, "inherit parent directory permission.")
	setACLCmd.Flags().StringP("owner", "o", "", "owner of the resource")
	setACLCmd.Flags().StringP("ownertype", "t", "0", "owner of the resource")
	fileCmd.AddCommand(setACLCmd)

	// set-ac
	setAc := &cobra.Command{
		Use:   "set-ac",
		Short: "Set the access control of file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("specify the resource path and recipient name")
			}
			path := args[0]
			recipient := args[1]
			acl, err := Client.GetACL(path)
			if err != nil {
				return err
			}
			delete, _ := cmd.Flags().GetBool("delete")
			if delete {
				acl.DeleteAC(recipient)
			} else {
				ac := acl.GetOrNewAC(recipient)
				if recipientType, _ := cmd.Flags().GetString("recipienttype"); recipientType != "" {
					ac.RecipientType = recipientType
				}
				if permissions, _ := cmd.Flags().GetString("permissions"); permissions != "" {
					ac.Permissions = permissions
				}
				if modifiable, _ := cmd.Flags().GetString("modifiable"); modifiable != "" {
					ac.Modifiable = modifiable
				}
				if err := acl.SetAC(ac); err != nil {
					return err
				}
			}
			return Client.SetACL(path, acl)
		},
	}
	setAc.Flags().StringP("recipienttype", "r", "", "Recipient type")
	setAc.Flags().StringP("permissions", "P", "", "Permission. (0:read,1:read/write,2:read/write/delete,4:read/write/delete/admin)")
	setAc.Flags().StringP("modifiable", "m", "", "Modifiable")
	setAc.Flags().BoolP("delete", "d", false, "Delete")

	fileCmd.AddCommand(&cobra.Command{
		Use:   "set-metadata",
		Short: "Set the metadata of files in repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return errors.New("specify the resource path, metadata key, and metadata value")
			}
			path := args[0]
			key := args[1]
			value := args[2]

			metadata, err := Client.GetMetadata(path)
			if err != nil {
				return err
			}

			metadata[key] = value
			return Client.SetMetadata(path, metadata)
		},
	})
	fileCmd.AddCommand(&cobra.Command{
		Use:   "get-metadata",
		Short: "get the metadata of files in repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the resource path")
			}
			path := args[0]
			metadata, err := Client.GetMetadata(path)
			if err != nil {
				return err
			}
			for key, value := range metadata {
				fmt.Printf("%s=%s\n", key, value)
			}
			return nil
		},
	})
}
