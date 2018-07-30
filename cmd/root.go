package cmd

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	client "github.com/uphy/pentahotools/client"

	"strings"

	"github.com/mattn/go-shellwords"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string
var url string
var user string
var password string

// Client for Pentaho
var Client client.Client

func initCommand(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
	})
	for _, c := range cmd.Commands() {
		initCommand(c)
	}
}

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
			splitedCommandLine, err := shellwords.Parse(commandLine)
			if err != nil {
				fmt.Println(err)
				continue
			}
			initCommand(cmd)
			cmd.SetArgs(splitedCommandLine)
			cmd.Execute()
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if Client.Logger != nil {
			Client.Logger.Error("Failed to execute command.", zap.String("error", err.Error()))
		}
		fmt.Println(err)
		os.Exit(1)
	} else {
		if Client.Logger != nil {
			Client.Logger.Debug("Successfully executed command.")
		}
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
