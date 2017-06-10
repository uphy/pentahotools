package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	client "github.com/uphy/pentahotools/client"

	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var url string
var user string
var password string

// Client for Pentaho
var Client client.Client

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "pentahotools",
	Short: "Pentaho CLI Tools",
	Long:  `Manage users, PDI jobs/transformations, repositositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Entering multiple command mode.")
		fmt.Println("Input 'exit' to exit this command.")
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			commandLine, _ := reader.ReadString('\n')
			commandLine = strings.TrimRight(commandLine, "\r\n ")
			if len(commandLine) == 0 {
				cmd.Help()
				continue
			}
			if commandLine == "exit" {
				break
			}
			splitedCommandLine := strings.Split(commandLine, " ")
			subProcessArgs := append(os.Args, splitedCommandLine...)
			subProcessCommand := exec.Command(subProcessArgs[0], subProcessArgs[1:]...)
			stdout, err := subProcessCommand.StdoutPipe()
			if err != nil {
				fmt.Println(err)
				continue
			}
			subProcessCommand.Start()

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Print(scanner.Text())
				fmt.Println()
			}

			subProcessCommand.Wait()
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initialize)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pentahotools.yaml)")
	RootCmd.PersistentFlags().StringVarP(&url, "url", "l", "http://localhost:8080/pentaho", "rest endpoint url")
	RootCmd.PersistentFlags().StringVarP(&user, "user", "u", "admin", "rest endpoint user name")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "password", "rest endpoint password")
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initialize() {
	url = strings.TrimSuffix(url, "/")
	Client = client.NewClient(url, user, password)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pentahotools" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pentahotools")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
