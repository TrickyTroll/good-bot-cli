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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// echoConfigCmd represents the echoConfig command
var echoConfigCmd = &cobra.Command{
	Use:   "echoConfig",
	Short: "Prints the parsed configuration file.",
	Long: `Can be used to quickly glance at how this
	program interprets your configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		echo()
	},
}

func init() {
	rootCmd.AddCommand(echoConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// echoConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// echoConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// echo gets the values that are set in a user's configuration file and prints
// them. If no value for a certain param can be found, the user is told that
// no value could be found.
func echo() {
	if viper.GetString("ttsCredentials") != "" {
		ttsCredentials := viper.GetString("ttsCredentials")
		fmt.Printf("Will be using TTS credentials from %s\n", ttsCredentials)
	} else {
		fmt.Println("There is no 'ttsCredentials' variable in your configuration file")
	}
	if viper.GetString("passwordsEnv") != "" {
		passwordsEnv := viper.GetString("passwordsEnv")
		fmt.Printf("Will be using passwords from env file %s\n", passwordsEnv)
	} else {
		fmt.Println("There is no 'passwordsEnv' variable in your configuration file")
	}
}
