package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var carteCmd = &cobra.Command{
	Use:   "carte",
	Short: "PDI operation command.",
	Long:  `Perform PDI operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var carteStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the carte server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := Client.GetCarteServerStatus()
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
	},
}

func init() {
	RootCmd.AddCommand(carteCmd)

	carteCmd.AddCommand(carteStatusCmd)
}
