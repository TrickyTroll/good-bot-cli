/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the app by pulling the latest docker image.",
	Long: `Updates the Good-Bot and the Asciicast2gif docker images.

It uses the equivalent of the docker pull command to update your
application. To see what will change, please refer to Good-Bot's
changelog.

It updates every image if no argument is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating images...")
		if len(args) > 0 {
			update(args)
		} else {
			defaultArgs := []string{"trickytroll/good-bot:latest", "asciinema/asciicast2gif"}
			update(defaultArgs)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func update(toUpdate []string) {

	for _, imageName := range toUpdate {

		fmt.Printf("Updating %s\n", imageName)
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil { // cli fails nothing else will work. Should panic.
			panic(err)
		}

		reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil { // If no reader the rest of the program won't work.
			panic(err)
		}
		io.Copy(os.Stdout, reader) // Print container info to stdout.
	}
}
