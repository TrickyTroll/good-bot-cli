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
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record [path to project]",
	Short: "Record a video using good-bot.",
	Long: `This command uses good-bot's container image
to record a project.

If the provided argument is a script, this command will
run the setup command first, and then use record.

If the argument is already a directory created by the
setup command, this command will only use the record
command to create the recordings.`,
	Run: func(cmd *cobra.Command, args []string) {
		credentials := copyCredentials()
		if isDirectory(args[0]) {
			runRecordCommand(args[0], credentials.ttsFile, credentials.passwords)
			if !noRender {
				renderProject(args[0])
			}
		} else {
			// TODO: Fix this.
			// runSetupCommand(args[0], "/project")
			// recDir := askRecDir()
			// runRecordCommand(recDir, credentials.ttsFile, credentials.passwords)
			// if !noRender {
			// 	renderProject("Toto")
			// }
			fmt.Printf("Cannot create a video from %s.\n", args[0])
			fmt.Println("Please make sure that you used the setup command first.")
			os.Exit(1)
		}
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument")
		} else if len(args) > 1 {
			return errors.New("requires at most one argument")
		} else if !validatePath(args[0]) { // d
			return errors.New("not a valid path")
		} else {
			return nil
		}
	},
}

type credentials struct {
	passwords []string
	ttsFile   string
}

var noRender bool

func init() {

	rootCmd.AddCommand(recordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	recordCmd.Flags().BoolVar(&noRender, "no-render", false, `If not rendering, Good-Bot only outputs asciicasts and 
audio recordings. No gifs or mp4 files are produced.`)
}

func runRecordCommand(hostPath string, ttsFile string, envVars []string) {
	// Used later for i/o between container and shell
	inout := make(chan []byte)

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

	ttyFileStats, err := os.Stat(ttsFile)
	if err != nil {
		panic(err)
	}
	ttsFileName := ttyFileStats.Name()

	containerTtsPath := "/credentials/" + ttsFileName

	credentialsEnv := fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", containerTtsPath)
	envVars = append(envVars, credentialsEnv)

	stats, err := os.Stat(hostPath)
	if err != nil {
		panic(err)
	}

	projectName := stats.Name()

	containerProjectPath := "/project" + "/" + projectName

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Env:          envVars,
		Cmd:          []string{"record", containerProjectPath},
		Image:        "trickytroll/good-bot:latest",
		Volumes:      map[string]struct{}{},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: getDir(hostPath),
				Target: "/project",
			},
			{
				Type:   mount.TypeBind,
				Source: getDir(ttsFile),
				Target: "/credentials",
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

// Path should be valid since it's been checked by validatePath()
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return info.IsDir()
}

func parsePasswords(passwordsPath string) ([]string, error) {
	file, err := os.Open(passwordsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.Trim(scanner.Text(), "\n"))
	}
	return lines, scanner.Err()
}

func copyCredentials() *credentials {

	ttsCredentials := viper.GetString("ttsCredentials")
	passwordsEnv := viper.GetString("passwordsEnv")

	allPasswords, _ := parsePasswords(passwordsEnv)

	return &credentials{allPasswords, ttsCredentials}
}
