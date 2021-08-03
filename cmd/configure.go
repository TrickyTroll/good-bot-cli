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

	"github.com/AlecAivazis/survey/v2"
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

// credentialsPaths is used to store results from getCredentials. The
// user is prompted for a path to his or her TTS authentication file
// and a path towards a passwords environment file.
type credentialsPaths struct {
	Tts string `survey:"tts"`
	Pass string `survey:"passwords"`
}

// setConfig creates an interactive prompt for the user to create a
// configuration file. This function uses promptCredentials to ask for
// paths towards the TTS API key and passwords file. The information
// collected by setConfig is then written to the chosen Viper
// configuration file.
func setConfig() {

	answers, err := promptCredentials()
	if err != nil {
		log.Fatal(err)
	}

	// Making sure that the user did provide a value.
	if len(answers.Tts) > 0 {
		absApiKeyPath, err := filepath.Abs(answers.Tts)
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("ttsCredentials", absApiKeyPath)
	}

	if len(answers.Pass) > 0 {
		absEnvFilePath, err := filepath.Abs(answers.Pass)
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("passwordsEnv", absEnvFilePath)
	}

	home, err := homedir.Dir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".good-bot-cli")
	viper.SetConfigType("yaml")

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

// promptCredentials prompts the user for paths towards their TTS
// credentials file and passwords environment file. The results are
// returned as a credentialsPaths structure.
//
// If an error is raised during prompting, the error is returned
// with a credentialsPaths with nothing defined at the Tts or Pass
// keys.
func promptCredentials() (credentialsPaths, error) {
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

	var answers credentialsPaths

	var qs = []*survey.Question {
		{
			Name: "tts",
			Prompt: &survey.Input {
				Message: "Please provide a path towards your Text-to-Speech API key",
				Help: "This can be an absolute or relative path.",
			},
			Validate: func (val interface{}) error {
				if str, ok := val.(string); !ok || !validatePath(str) {
					// Accepting empty values
					if len(str) == 0 {
						return nil
					}
					return errors.New("the path provided does not seem to be valid")
				}
				return nil
			},
		},
		{
			Name: "passwords",
			Prompt: &survey.Input {
				Message: "Please provide a path towards your passwords environment file",
				Help: "This can be an absolute or relative path.",
			},

			Validate: func (val interface{}) error {
				if str, ok := val.(string); !ok || !validatePath(str) {
					// Accepting empty values
					if len(str) == 0 {
						return nil
					}
					return errors.New("the path provided does not seem to be valid")
				}
				return nil
			},
		},
	}
	err := survey.Ask(qs, &answers)

	if err != nil {
		return answers, err
	}
	return answers, nil
}
