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
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
			fmt.Println("Ok. Setting up your configuration file now.")
			setConfig()
		} else {
			fmt.Println("Ok. Won't be setting a configuration file for now.")
		}
	}
}

// validatePath checks whether or not a path exists. The check is done using
// Stat on the path. If there is no error using Stat, validatePath returns
// true, else it returns false.
func validatePath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getDir gets the directory where a file is saved. The path returned by
// this function is a full path. If the current working directory cannot
// be found, filepath.Abs returns an error. This error is handled by a
// panic.
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

// dockerCheck checks whether or not the user has Docker installed and
// available. If Docker cannot be found from exec.LookPath, the program
// exits using log.Fatal.
func dockerCheck() {
	_, err := exec.LookPath("docker")
	if err != nil {
		log.Fatal(err)
	}
}

// askSetConfig prompts the user on whether or not the CLI should be configured
// right now.
//
// It uses the survey library to provide an interactive yes/no prompt.
//
// The result is then returned as a bool (true for yes false for no).
func askSetConfig() bool {
	fmt.Println("No configuration file was found!")
	var setConfig bool
	prompt := &survey.Confirm{
		Message: "Would you like to create one now?",
	}
	err := survey.AskOne(prompt, &setConfig)
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return setConfig
}

// isReadStatement reads a script file line by line and checks whether
// or not a read statement is used. To make this check, each scene in
// the project directory is searched. If there is a non-empty directory
// called "read" in one of the scenes, this function returns true.
func isReadStatement(scriptPath string) (bool, error) {
	file, err := os.Open(scriptPath)
	defer file.Close()

	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		contents := scanner.Text()
		if strings.Contains(contents, "read:") || strings.Contains(contents, "read :") {
			return true, nil
		}
	}

	// File is closed using the defer statement
	return false, nil
}
