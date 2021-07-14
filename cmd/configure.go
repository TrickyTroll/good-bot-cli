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
	"path/filepath"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up a configuration file interactively.",
	Long: `Sets up the configuration file interactively by asking
where some required files are located on the host computer.

Good-bot will use this information to mount the files'
locations when running the Docker container.`,
	Run: func(cmd *cobra.Command, args []string) {
		setConfig()
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// setConfig creates an interactive prompt for the user to create a
// configuration file. This function uses promptui to ask for a path
// towards the TTS API key and a path towards the passwords file. The
// information collected by setConfig is then written to the chosen
// Viper configuration file.
func setConfig() {
	/*
		There are 2 things that a user needs to set up in
		order to user good-bot-cli.

		1. A path towards their TTS api key.
		2. A path towards their  password env file.

		Since those are paths, they are verified with the
		validate function below to make sure that

		* They exist
		* They are accessible from the user that is running
		  this program.
	*/
	validatePath := func(path string) error {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return errors.New("file does not exist")
		}
		return nil
	}

	promptApiKey := promptui.Prompt{
		Label:    "Please provide a path towards your Text-to-Speech API key",
		Validate: validatePath,
	}

	promptEnvFile := promptui.Prompt{
		Label:    "Please provide a path towards your passwords environment file",
		Validate: validatePath,
	}

	apiKeyPath, err := promptApiKey.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	envFilePath, err := promptEnvFile.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	absApiKeyPath, _ := filepath.Abs(apiKeyPath)
	absEnvFilePath, _ := filepath.Abs(envFilePath)

	home, err := homedir.Dir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".good-bot-cli")
	viper.SetConfigType("yaml")
	viper.Set("ttsCredentials", absApiKeyPath)
	viper.Set("passwordsEnv", absEnvFilePath)

	file, err := os.Create(home + "/.good-bot-cli.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = viper.WriteConfig()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Configuration file has been written as %s.\n", viper.ConfigFileUsed())
}
