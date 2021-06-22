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
)

const recordingsPath string = "/asciicasts"
const renderPath string = "/gifs"

func renderProject(projectPath string) {
	toRecord := getRecsPaths(projectPath)
	fmt.Println("Rendering...")
	fmt.Println(projectPath)
	fmt.Println(toRecord)
	for _, item := range toRecord {
		renderRecording(item, projectPath)
	}
}

func renderRecording(asciicastPath, projectPath string) string {

	stats, err := os.Stat(asciicastPath)

	if err != nil {
		// Providing a path to an asciicast that does not exists to
		// this function should never happen.
		log.Panic(err)
	}

	splitProjectName := strings.Split(projectPath, "/")
	projectName := splitProjectName[len(splitProjectName)-1]
	noExt := strings.TrimSuffix(asciicastPath, filepath.Ext(asciicastPath))

	// The project path is mounted as `/data` in the container.
	recContainerPath := "/data/" + projectName + recordingsPath + stats.Name()
	gifContainerPath := "/data/" + projectName + renderPath + noExt + ".gif"

	// Used later for i/o between container and shell
	inout := make(chan []byte)

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

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Cmd:   []string{recContainerPath, gifContainerPath},
		Image: "asciinema/asciicast2gif",
	}, &container.HostConfig{
		Mounts: []mount.Mount{ // Mounting the location where the script is written.
			{
				Type:   mount.TypeBind,
				Source: getDir(projectPath), // `getDir()` defined in `root.go`.
				Target: "/data",             // Specified in asciicast2gif's README.
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

	return projectName + renderPath + "/" + noExt + ".gif"
}

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
		fmt.Println(dir)
		sceneRecordings := getSceneCasts(scenePath)
		allPaths = append(allPaths, sceneRecordings...)
	}
	return allPaths
}

func getSceneCasts(scenePath string) []string {
	var allScenePaths []string
	castsPath := scenePath + recordingsPath
	recordings, err := ioutil.ReadDir(castsPath)
	if err != nil {
		// Probabably means that there are no recordings. No need to Panic.
		log.Println(fmt.Sprintf("Found no recordings in scene %s", scenePath))
		return nil
	}
	for _, file := range recordings {
		fmt.Println(file)
		filePath := castsPath + "/" + file.Name()
		fmt.Println(filepath.Ext(filePath))
		if filepath.Ext(filePath) == ".cast" {
			fmt.Println("Found an asciicast!")
			allScenePaths = append(allScenePaths, filePath)
		}
	}
	return allScenePaths
}
