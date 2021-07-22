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
	"encoding/json"
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
		dockerCheck()
		// First argument should be the project path.
		renderAllRecordings(args[0])
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

// Settings that should be found in an asciicast v2 file.
type asciicastSettings struct {
	Version int `json:"version"`
	Width   int `json:"width"`
	Height  int `json:"height"`
	Time    int `json:"timestamp"`
	Env     struct {
		Shell bool   `json:"SHELL"`
		Term  string `json:"TERM"`
	}
}

// renderAllRecordings uses renderRecording on each Asciinema recording from
// a project. It uses getRecsPaths to get an array of paths towards each
// asciicast. This function also pulls Asciicast2gif's Docker image, and
// provides a client and context  to renderRecording.
func renderAllRecordings(projectPath string) {
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
		renderRecording(item, cli, ctx)
	}
}

// renderRecording uses Asciicast2gif's Docker image to convert an
// asciicast to the gif format. This function does not pull the
// Docker image, so it needs the client and context passed as arguments.
// Asciicast2gif is used with the "-S1" flag to reduce the gif's
// resolution.
func renderRecording(asciicastPath string, cli *client.Client, ctx context.Context) {

	stat, err := os.Stat(asciicastPath)

	if err != nil {
		log.Fatal(err)
	}

	// Cropping to 24x80
	// cropRec(asciicastPath)

	fileName := strings.TrimSuffix(stat.Name(), filepath.Ext(stat.Name()))

	scenePath, err := getScenePath(asciicastPath)

	if err != nil {
		log.Printf("Could not render file: %s\n%s", asciicastPath, err)
		return
	}

	outputPath := filepath.Join(scenePath, renderPath, fileName+".gif")

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

// renderVideo uses Good Bot's Docker image to render a previously
// recorded video. It uses the render-video command. The project path
// is used to mount the project's location to the container, since
// the commands needs access to the project.
//
// The Docker image is pulled on each run. The project path is
// mounted under /project/[PROJECT NAME] in the container.
//
// The conversion from Asciinema recordings to the gif format
// is not done with Good Bot's Docker image. good-bot-cli
// instead uses the Asciicast2gif image. The conversion of
// every asciicast is done with the renderRecordings function.
//
// The final video is written in the projectPath/final directory.
//
// This function also returns the final video's path.
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

	containerProjectPath := filepath.Join("/project", projectName)
	finalPath := filepath.Join(projectPath, "/final")

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

// cropRec "crops" an Asciinema recording to the standard 24x80
// format. The "cropping" is done by changing the width and height
// parameters from the asciicast's json recording file.
func cropRec(recPath string) error {

	var linesBytes [][]byte

	file, err := os.Open(recPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		linesBytes = append(linesBytes, scanner.Bytes())
	}

	config, err := getAsciicastConfig(linesBytes)

	if err != nil {
		return err
	}

	config.Height = 80
	config.Width = 24

	newFirstLine, err := json.Marshal(config)

}

// getAsciicastConfig gets an asciicast's configuration information.
// The settings are unmarshalled from the asciicast's first line.
//
// The settings are unmarshalled in a struct type defined as
// asciicastSettings.
//
// An error is returned if there was an error returned by os.Open or
// json.Unmarshal.
func getAsciicastConfig(fileLines [][]byte) (*asciicastSettings, error) {

	var settings asciicastSettings

	err := json.Unmarshal(fileLines[0], &settings)

	if err != nil {
		return nil, err
	}

	return &settings, err
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
		scenePath := filepath.Join(projectPath, dir.Name())
		sceneRecordings := getSceneCasts(scenePath)
		allPaths = append(allPaths, sceneRecordings...)
	}
	return allPaths
}

// getSceneCasts looks for each Asciinema recording saved under
// the provided scene path. For each file contained in the
// recordings path of a scene, this function checks if the file's
// extension is ".cast". Each match is appended to an array of
// paths which is then returned.
func getSceneCasts(scenePath string) []string {
	var sceneRecordings []string
	castsPath := filepath.Join(scenePath, recordingsPath)
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
// from the filepath library until the base of the path contains
// "scene_".
//
// Returns an absolute path.
func getScenePath(itemPath string) (string, error) {

	// scenePath initially not really the scene's path
	scenePath := itemPath

	// Cheking if file exists.
	_, err := os.Stat(itemPath)
	if err != nil {
		return "", err
	}

	base := filepath.Base(itemPath)
	// Go back until the base of the path  contains "scene_"
	for {
		if strings.Contains(base, "scene_") {
			break
		}
		scenePath = filepath.Dir(scenePath)
		base = filepath.Base(scenePath)
		if strings.HasSuffix(scenePath, string(os.PathSeparator)) {
			err := fmt.Errorf("%s does not seem to be saved in a Good Bot project", itemPath)
			return "", err
		}
	}
	// The scene path should be the parent dir
	// of the type of media.
	abs, err := filepath.Abs(scenePath)

	if err != nil {
		return "", err
	}

	return abs, nil
}
