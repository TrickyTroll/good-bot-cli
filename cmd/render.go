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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render [path to project]",
	Short: "Renders a previously recorded project.",
	Long: `Render can be used to create a video from a recording
that used the --no-render flag.

The --no-render flag can be used to speed up the recording
process. Once you are happy with the result, the video can be
rendered afterwards using this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// First argument should be the project path.
		renderProject(args[0])
		renderVideo(args[0])
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
	rootCmd.AddCommand(renderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// renderCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// renderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const recordingsPath string = "/asciicasts/"
const renderPath string = "/gifs/"

func renderProject(projectPath string) {
	// Spawning it only once
	// Normal context with no timeout.
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil { // cli fails nothing else will work. Should panic.
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "asciinema/asciicast2gif", types.ImagePullOptions{})
	if err != nil { // If no reader the rest of the program won't work.
		panic(err)
	}
	io.Copy(os.Stdout, reader) // Print container info to stdout.

	toRecord := getRecsPaths(projectPath)
	for _, item := range toRecord {
		renderRecording(item, projectPath, cli, ctx)
	}
}

func renderRecording(asciicastPath string, projectPath string, cli *client.Client, ctx context.Context) {

	stat, err := os.Stat(asciicastPath)

	if err != nil {
		log.Fatal(err)
	}

	// Cropping to 24x80
	// cropRec(asciicastPath)

	fileName := strings.TrimSuffix(stat.Name(), filepath.Ext(stat.Name()))

	scenePath := getScenePath(asciicastPath)

	outputPath := scenePath + renderPath + fileName + ".gif"

	currentWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Used later for i/o between container and shell
	inout := make(chan []byte)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Cmd:   []string{"-S1", asciicastPath, outputPath},
		Image: "asciinema/asciicast2gif",
	}, &container.HostConfig{
		Mounts: []mount.Mount{ // Mounting the location where the script is written.
			{
				Type:   mount.TypeBind,
				Source: currentWorkingDir, // `getDir()` defined in `root.go`.
				Target: "/data",           // Specified in asciicast2gif's README.
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

func renderVideo(projectPath string) string {
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

	stats, err := os.Stat(projectPath)
	if err != nil {
		panic(err)
	}

	projectName := stats.Name()

	containerProjectPath := "/project" + "/" + projectName
	finalPath := projectPath + "/final"

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		OpenStdin:    true,
		Cmd:          []string{"render-video", containerProjectPath},
		Image:        "trickytroll/good-bot:latest",
		Volumes:      map[string]struct{}{},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: getDir(projectPath),
				Target: "/project",
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

	return finalPath
}

func cropRec(recPath string) error {
	file, err := ioutil.ReadFile(recPath)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(file), "\n")
	params := strings.Split(lines[1], ",")
	newWidth := replaceParam(params[1], "24")
	newHeight := replaceParam(params[2], "80")

	params[1] = newWidth
	params[2] = newHeight

	lines[0] = strings.Join(params, ",")

	ioutil.WriteFile(recPath, []byte(strings.Join(lines, "\n")), 0644)

	return nil
}

// replaceParam replaces a value for a certain parameter from
// an Asciinema recording file. It returns the provided string
// with the new parameter instead  of the old one.
func replaceParam(paramString string, newParam string) string {
	editing := strings.Split(paramString, ":")
	editing[1] = " " + newParam

	return strings.Join(editing, ":")
}

// getRecsPaths fetches every recording for a project.
// It gets a list of every scene and then it uses
// getSceneCasts to get each Asciinema recording from
// each scene. It returns an array of paths towards
// every recording saved in the provided project path.
func getRecsPaths(projectPath string) []string {
	var allPaths []string
	// Each dir is a `scene`.
	dirs, err := ioutil.ReadDir(projectPath)
	if err != nil {
		// The function should panic since providing a path
		// that does not exists here is unexpected.
		log.Panic(err)
	}
	for _, dir := range dirs {
		scenePath := projectPath + "/" + dir.Name()
		sceneRecordings := getSceneCasts(scenePath)
		allPaths = append(allPaths, sceneRecordings...)
	}
	return allPaths
}

// getSceneCasts looks for each Asciinema recording saved under
// the provided scene path. For each file contained in the
// recordings path of a scene, this function checks if the file's
// extension is ".cast". Each match is appended to an array of
// paths which is then returned..
func getSceneCasts(scenePath string) []string {
	var sceneRecordings []string
	castsPath := scenePath + recordingsPath
	recordings, err := ioutil.ReadDir(castsPath)
	if err != nil {
		// Probabably means that there are no recordings. No need to Panic.
		log.Println(fmt.Sprintf("Found no recordings in scene %s", scenePath))
		return nil
	}
	for _, file := range recordings {
		filePath := scenePath + recordingsPath + file.Name()
		if filepath.Ext(filePath) == ".cast" {
			sceneRecordings = append(sceneRecordings, filePath)
		}
	}
	return sceneRecordings
}

// getScenePath searches for the name of the scene where the
// provided recording path is saved. It uses the Dir method
// from the filepath library twice on the provided recording
// path.
func getScenePath(recPath string) string {
	typePath := filepath.Dir(recPath)
	// The scene path should be the parent dir
	// of the type of media.
	return filepath.Dir(typePath)
}
