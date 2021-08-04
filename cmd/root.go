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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
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
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		cobra.CheckErr(err)

		// Search config in home directory with name ".good-bot-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".good-bot-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match
}

// validatePath checks whether or not a path exists. The check is done using
// Stat on the path. If there is no error using Stat, validatePath returns
// true, else it returns false.
func validatePath(path string) bool {
	processed, err := processPath(path)
	// TODO: return the error in a next version of validatePath
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(processed)
	return err == nil
}

// processPath takes a path as an input and returns a path that
// other Go functions can use more easily.
//
// If the path starts with a "~" character, "~" is replaced by
// the full path to the user's home directory.
//
// If any error is encountered while trying to find the user's
// home directory, or while getting an absolute path, the error
// is returned.
//
// This function always returns an absolute path.
func processPath(path string) (string, error) {
	var processedPath string

	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		processedPath = filepath.Join(homeDir, path[1:])
	} else if strings.HasPrefix(path, ".") {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		processedPath = filepath.Join(currentDir, path[1:])
	} else {
		// This is in an else statement because the first if
		// guarantees an absolute path.
		var err error
		processedPath, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}
	return processedPath, nil
}

// mergeProcessedPaths merges prefix and path to create an absolute path from a
// relative path. prefix should be determined prior to running this function.
//
// If the length of path is smaller than 1, an empty string is returned along
// with an error.
func mergeProcessedPaths(prefix, path string) (string, error) {
	var joinedPaths string
	if len(path) >= 2 {
		joinedPaths = filepath.Join(prefix, path[1:])
	} else if len(path) == 1 {
		joinedPaths = filepath.Join(prefix, path)
	} else {
		err := fmt.Errorf("path '%s' provided was invalid, had length %d.", path, len(path))
		return "", err
	}
	return joinedPaths, nil
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

// setConfigInteraction checks if a configuration file exists. If it
// doesn't, the user will be prompted on whether or not an interactive
// configuration process should be started.
//
// If the user wants to configure the program now, the interactive
// prompting will begin and the config file will be filled.
//
// If the configuration is postponed, an empty configuration file will
// be created.
func setConfigInteraction() {
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Got an error trying to find the home directory.\n%s", err)
		}
		var configFile string = filepath.Join(homeDir, ".good-bot-cli.yaml")
		if askSetConfig() {
			fmt.Println("Ok. Setting up your configuration file now.")
			setConfig()
		} else {
			fmt.Println("Ok. Won't be setting a configuration file for now.")
			// Writing an empty line
			err = ioutil.WriteFile(configFile, []byte("\n"), 0644)
			if err != nil {
				log.Printf("Could not write an empty configuration file %s\n%s", configFile, err)
			}
		}
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
//
// If there is an error returned when reading a directory, it is
// returned and the boolean value returned is false.
func isReadStatement(projectPath string) (bool, error) {
	projectContents, err := os.ReadDir(projectPath)

	if err != nil {
		return false, err
	}

	for _, scene := range(projectContents) {
		// Ignoring items that aren't scenes.
		if strings.Contains(scene.Name(), "scene_") && scene.IsDir() {
			sceneContents, err := os.ReadDir(filepath.Join(projectPath, scene.Name()))

			if err != nil {
				return false, err
			}

			for _, item := range(sceneContents) {
				if strings.Contains(item.Name(), "read") && item.IsDir() {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
