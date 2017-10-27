package cmd

import (
	"fmt"
	"os"

	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	client "github.com/uphy/pentahotools/client"
)

func init() {
	var carteCmd = &cobra.Command{
		Use:   "carte",
		Short: "PDI operation command.",
		Long:  `Perform PDI operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	RootCmd.AddCommand(carteCmd)

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show status of the carte server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("too many arguments")
			}
			// Show job and trans list
			if len(args) == 0 {
				status, err := Client.GetStatusCarteServer()
				if err != nil {
					return err
				}
				status.SortStatusByLogDate()

				fmt.Println("# Environment Status")
				fmt.Printf("Status: %s\n", status.StatusDescription)
				fmt.Printf("Memory Usage: %3.1f/%3.1f MB (%3.1f %%)\n", float32(status.MemoryTotal-status.MemoryFree)/1024./1024., float32(status.MemoryTotal)/1024./1024., float64((status.MemoryTotal-status.MemoryFree))/float64(status.MemoryTotal)*100)
				fmt.Printf("CPU Cores: %d\n", status.CPUCores)
				fmt.Printf("CPU Process Time: %d\n", status.CPUProcessTime)
				fmt.Printf("Uptime: %d\n", status.UpTime)
				fmt.Printf("Thread Count: %d\n", status.ThreadCount)
				fmt.Printf("Load Average: %3.2f\n", status.LoadAverage)
				fmt.Printf("OS: %s %s %s\n", status.OSName, status.OSVersion, status.OSArch)
				fmt.Println()

				fmt.Println("# Job and Transformation Status")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Type", "Name", "Date", "ID", "Status", "Logging String"})
				for _, status := range status.JobStatusList.List {
					table.Append([]string{"Job", status.Name, status.LogDate, status.ID, status.StatusDescription, status.LoggingString})
				}
				for _, status := range status.TransformationStatusList.List {
					table.Append([]string{"Trans", status.Name, status.LogDate, status.ID, status.StatusDescription, status.LoggingString})
				}
				table.SetAutoMergeCells(true)
				table.SetRowLine(true)
				table.Render()
				return nil
			}
			// Show transformation status
			var status client.Status
			var err error
			id, name := client.ParseIDAndName(args[0])
			status, err = Client.GetStatus(id, name, 0)
			if err != nil {
				return err
			}
			status.Print(client.NewIndentWriter(os.Stdout))
			return nil
		},
	}
	statusCmd.Aliases = []string{"ls"}
	carteCmd.AddCommand(statusCmd)
	carteCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run the specified job or transformation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a job")
			}
			jobID, err := Client.Run(args[0], client.LogLevels.Debug)
			if err != nil {
				return errors.Wrap(err, "job execution failure")
			}
			status, err := Client.GetStatus(jobID, "", 0)
			if err != nil {
				return errors.Wrap(err, "getting status failure")
			}
			for {
				status.Print(client.NewIndentWriter(os.Stdout))
				if status.IsFinished() {
					break
				}
				time.Sleep(time.Second)
			}
			return nil
		},
	})

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove the specified job/transformation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all, _ := cmd.Flags().GetBool("all"); all {
				status, err := Client.GetStatusCarteServer()
				if err != nil {
					return errors.Wrap(err, "getting job list failure")
				}
				for _, job := range status.JobStatusList.List {
					err = Client.JobClient.Remove(job.ID, job.Name)
					if err != nil {
						return errors.Wrap(err, "job removal failure")
					}
				}
				for _, trans := range status.TransformationStatusList.List {
					err = Client.TransformationClient.Remove(trans.ID, trans.Name)
					if err != nil {
						return errors.Wrap(err, "transformation removal failure")
					}
				}
			} else {
				if len(args) != 1 {
					return errors.New("specify a job or transformation")
				}
				var err error
				id, name := client.ParseIDAndName(args[0])
				err = Client.RemoveJobOrTransformation(id, name)
				if err != nil {
					return errors.Wrap(err, "job/transformation removal failure")
				}
			}
			return nil
		},
	}
	removeCmd.Flags().BoolP("all", "a", false, "Remove all finished job/transformations.")
	removeCmd.Aliases = []string{"rm"}
	carteCmd.AddCommand(removeCmd)
}
