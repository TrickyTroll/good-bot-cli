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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Sets up your project directory from a configuration file.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSetupCommand(args[0], "/project")
	},
	// Makes sure that there is one argument and it is an
	// existing file.
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument")
		} else if len(args) > 1 {
			return errors.New("requires at most one argument")
		} else if !validatePath(args[0]) {
			return errors.New("file does not exist")
		} else {
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

func runSetupCommand(filePath string, containerPath string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "trickytroll/good-bot:latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	// Script and infos are written in containerPath. The directory
	// where the script resides on the host will be mounted to containerPath.
	stats, err := os.Stat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	scriptName := stats.Name()

	containerScriptPath := containerPath + "/" + scriptName

	resp, err := cli.ContainerCreate(ctx, &container.Config{

		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		ExposedPorts: map[nat.Port]struct{}{},
		Tty:          true,
		OpenStdin:    true,
		StdinOnce:    false,
		Cmd:          []string{"setup", containerScriptPath},
		Image:        "trickytroll/good-bot:latest",
		Volumes:      map[string]struct{}{},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: getDir(filePath),
				Target: containerPath,
			},
		},
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}
	fmt.Println("Created container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	fmt.Println("Started container")
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
