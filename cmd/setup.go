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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup [path to script]",
	Short: "Sets up your project directory from a configuration file.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCheck()
		runSetupCommand(args[0], "/project")
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument")
		} else if len(args) > 1 {
			return errors.New("requires at most one argument")
		} else if !validatePath(args[0]) {
			return errors.New("not a valid path")
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

// projectSaveInfo is used ton store results from getProjectPath. The
// user is prompted for a save path and a file name, which are stored as
// "path" and "name", respectively.
type projectSaveInfo struct {
		Path string `survey:"path"`
		Name string `survey:"name"`
}

// runSetupCommand uses Good Bot's Docker image to set up the project. It pulls
// the image on each run. The container's output is copied to the shell's
// stdout and the container is started interactively. This allows the user to
// answer the prompt when Good Bot asks for the name of the project.
//
// filePath is the path towards the script file, and containerPath is the path
// towards the file once it is mounted in the container.
//
// This function also uses the Docker SDK to mount the directory where the
// configuration file is located.This is also where the project directory
// will be created.
func runSetupCommand(filePath string, containerPath string) {

	// Used later for i/o between container and shell
	inout := make(chan []byte)

	// Normal context with no timeout.
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil { // cli fails nothing else will work. Should panic.
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "trickytroll/good-bot:latest", types.ImagePullOptions{})
	if err != nil { // If no reader the rest of the program won't work.
		panic(err)
	}
	io.Copy(os.Stdout, reader) // Print container info to stdout.

	// Script and infos are written in containerPath. The directory
	// where the script resides on the host will be mounted to containerPath.
	stats, err := os.Stat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	scriptName := stats.Name()

	containerScriptPath := containerPath + "/" + scriptName
	writeLoc := "/users-cwd"
	projectPath, err := getProjectPath()

	if err != nil {
		log.Fatal(err)
	}

	containerWritePath := filepath.Join(writeLoc, projectPath.Name)
	hostWritePath, err := filepath.Abs(projectPath.Path)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{

		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Cmd:          []string{"setup", "--project-path",containerWritePath, containerScriptPath},
		Image:        "trickytroll/good-bot:latest",
	}, &container.HostConfig{
		Mounts: []mount.Mount{ // Mounting the location where the script is written.
			// Mounting the location of the config file.
			{
				Type:   mount.TypeBind,
				Source: getDir(filePath),
				Target: containerPath,
			},
			// Mounting the write location of the project directory.
			{
				Type: mount.TypeBind,
				Source: hostWritePath,
				Target: writeLoc,
			},
		},
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// Need to attach since the user will be interacting with the container
	waiter, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})

	// Starting a goroutine for copying. Copies container output to stdout.
	go io.Copy(os.Stdout, waiter.Reader)
	go io.Copy(os.Stderr, waiter.Reader)

	if err != nil {
		panic(err)
	}

	go func() { // In a goroutine
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() { // Write terminal input to inout channel
			inout <- []byte(scanner.Text())
		}
	}()

	go func(w io.WriteCloser) { // In another goroutine
		for {
			data, ok := <-inout // Get terminal input from channel
			if !ok {
				fmt.Println("!ok")
				w.Close()
				return
			}

			w.Write(append(data, '\n')) // Write input to `w`. `w` is a Conn interface.
			// See https://pkg.go.dev/net#Conn
		}
	}(waiter.Conn)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case status := <-statusCh:
		fmt.Printf("status.StatusCode: %#+v\n", status.StatusCode)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

// getProjectPath prompts the user for a project save path and a project
// name. The path and the name are then joined to return the path towards
// where the new project should be written.
//
// An error is returned if it is encountered while prompting.
func getProjectPath() (projectSaveInfo, error) {

	var answers projectSaveInfo

	var qs = []*survey.Question{
		{
			Name: "path",
			Prompt: &survey.Input{
				Message: "Where do you want to save your project?",
				Help: "Provide an existing directory on your system. Good Bot will write your\nproject configuration in this directory.",
			},
			// Making sure that the directory exists
			Validate: func (val interface{}) error {
				if str, ok := val.(string); !ok || !validatePath(str) {
					return errors.New("the path provided does not seem to be valid")
				}
				return nil
			},
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "How do you want to name your project?",
				Help: "Provide a name for your project. The configuration directory will be saved\nunder the path:[project path]/[project name]",
			},
			Validate: func (val interface {}) error {
				str, ok := val.(string)
				if !ok {
					return errors.New("could not use your path as a string")
				}
				if strings.Contains(str, string(os.PathSeparator)) {
					return errors.New("your project name cannot contain path separators")
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
