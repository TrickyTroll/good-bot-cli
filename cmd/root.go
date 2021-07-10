/*
Copyright © 2021 Etienne Parent <tricky@beon.ca>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "good-bot-cli",
	Short: "Good Bot's command line interface.",
	Long: `good-bot-cli makes it easier to use good-bot in a container.

It offers the same commands as the Docker application but with a more intuitive
CLI.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.good-bot-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func validatePath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getDir(path string) string {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		// This should never happen since the path is checked
		// using validatePath. Panic if it happens.
		panic(err)
	}
	fileDir := filepath.Dir(fullPath)

	return fileDir
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".good-bot-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".good-bot-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if askSetConfig() {
			setConfig()
		} else {
			fmt.Println("Ok. Won't be setting a configuration file for now.")
		}
	}
}

func dockerCheck() {
	_, err := exec.LookPath("docker")
	if err != nil {
		log.Fatal(err)
	}
}

func askRecDir() string {
	validatePath := func(path string) error {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return errors.New("file does not exist")
		}
		return nil
	}

	promptRecDir := promptui.Prompt{
		Label:    "What is the path towards the project you want to record?\n",
		Validate: validatePath,
	}

	recDir, err := promptRecDir.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return recDir
}

func askSetConfig() bool {
	fmt.Println("No configuration file was found!")
	fmt.Println("Would you like to create one now (yes/no)? ")
	prompt := promptui.Select{
		Label: "Select[Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}
