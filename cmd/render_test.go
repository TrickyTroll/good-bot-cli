package cmd

import (
	"bufio"
	"io/ioutil"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"testing"
	"context"
)

func TestRenderRecording(t *testing.T) {

	asciicastPath, err := filepath.Abs("../testdata/scene_1/asciicasts/commands_1.cast")

	if err != nil {
		t.Errorf("Test error: could not find absolute path in TestRenderRecording.\n%s", err)
	}

	_, err = os.Stat(asciicastPath)

	if err != nil {
		t.Errorf("Test error: file provided in TestRenderRecording does not seem to be valid.\n%s", err)
	}
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

	render := renderRecording(asciicastPath, cli, ctx)

	// Checking if file has been properly created.
	_, err = os.Stat(render)

	if err != nil {
		t.Errorf("renderRecording on file %s did not produce a valid outpuput.\nCalling os.Stat on the file created by renderRecording returned error:\n%s", asciicastPath, err)
	}

	// Cleaning up
	os.Remove(render)
}

// TestCropRec creates a copy of an existing recording with wrong
// height and width to test the cropRec function. It checks wether
// or not the cropped recording is in the 24x80 format.
func TestCropRec(t *testing.T) {
	// Creating copy of test file
	castPath, err := filepath.Abs("../testdata/croptests/commands_1.cast")

	if err != nil {
		t.Errorf("Test error: Error finding testdata: %s\n", err)
	}

	newCastPath, err := filepath.Abs(filepath.Join(filepath.Dir(castPath), "castCopy.yaml"))

	contents, err := os.ReadFile(castPath)

	if err != nil {
		t.Errorf("Test error: could not read file %s.\n%s", castPath, err)
	}

	err = ioutil.WriteFile(newCastPath, contents, 0644)
	defer os.Remove(newCastPath)

	if err != nil {
		t.Errorf("Test  error: could not write to file.\n%s", err)
	}

	// Cropping the new test file.
	cropRec(newCastPath)

	// Getting lines from new file
	file, err := os.Open(newCastPath)
	if err != nil {
		t.Errorf("Test error: could not open file %s\n%s", newCastPath, err)
	}
	defer file.Close()

	var fileLines [][]byte

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Bytes())
	}

	info, err := getAsciicastConfig(fileLines)

	if err != nil {
		t.Errorf("getAsciicastConfig running on %s and got error:\n%s", newCastPath, err)
	}

	if info.Height != 24 {
		t.Errorf("cropRec error: cropped file %s has height %d, want %d", newCastPath, info.Height, 24)
	}

	if info.Width != 80 {
		t.Errorf("cropRec error: cropped file %s has width %d, want %d", newCastPath, info.Width, 80)
	}

	// Files are closed with defer statements.
}

// TestGetRecPaths checks the amount of asciicasts found in a project
// by getRecsPaths. The project used for those tests contains dummy
// files in one of the scene's asciicast directory.
func TestGetRecsPaths(t *testing.T) {

	projectPath, err := filepath.Abs("../testdata")

	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}

	recPaths := getRecsPaths(projectPath)

	// There should be 5 asciicasts in the project
	want := 5
	got := len(recPaths)

	if got != want {
		t.Errorf("getRecsPaths(%s) returns an array of length %d, want %d", projectPath, got, want)
	}
}

// TestGetSceneCastsContents makes sure that the contents of each
// file that corresponds to a path returned by getSceneCasts seems
// to be a valid asciicast. This test is made by checking if the
// contents of the file's first line contains the same keys as an
// asciicast v2 file.
func TestGetSceneCastsContents(t *testing.T) {

	scenePath, err := filepath.Abs("../testdata/scene_1")

	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}

	allFiles, err := ioutil.ReadDir(filepath.Join(scenePath, recordingsPath))

	if err != nil {
		t.Error(err)
	}

	want := 2

	if len(allFiles) != want {
		t.Errorf("testdata in %s should contain %d asciicasts", scenePath, want)
	}

	casts, err := getSceneCasts(scenePath)

	if err != nil {
		t.Errorf("getSceneCasts(%s) returned an error:\n%s", scenePath, err)
	}

	// Checking contents of each file.
	for _, file := range casts {
		var fileLines [][]byte

		openFile, err := os.Open(file)

		if err != nil {
			t.Errorf("test: error reading file %s.\n%s", file, err)
		}

		scanner := bufio.NewScanner(openFile)

		for scanner.Scan() {
			fileLines = append(fileLines, scanner.Bytes())
		}

		_, err = getAsciicastConfig(fileLines)

		if err != nil {
			t.Errorf("parsing json from asciicast %s returned an error:\n%s\n", file, err)
		}
	}
}

// TestGetAsciicastConfig makes sure that the configuration returned by
// getAsciicastConfig contains the right parameters. The asciicast used
// for this test is commands_1.cast in the testdata/croptests.
func TestGetAsciicastConfig(t *testing.T) {

	recPath, err := filepath.Abs("../testdata/croptests/commands_1.cast")

	if err != nil {
		t.Errorf("Error finding file: %s", err)
	}

	var fileLines [][]byte

	openFile, err := os.Open(recPath)

	if err != nil {
		t.Errorf("test: error reading file %s.\n%s", recPath, err)
	}

	scanner := bufio.NewScanner(openFile)

	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Bytes())
	}

	recSettings, err := getAsciicastConfig(fileLines)

	// Checking each param in the config. Th json looks
	// like this:
	// {
	// 	"version": 2,
	// 	"width: 219,
	// 	"height": 8,
	// 	"timestamp": 1625778960,
	// 	"env": {"SHELL": null, "TERM": "linux"}
	// }
	wantSettings := &asciicastSettings{2, 219, 8, 1625778960, &asciicastEnv{"", "linux"}}

	if recSettings.Version != wantSettings.Version {
		t.Errorf("getAsciicastConfig on file %s found version %d, want %d.", recPath, recSettings.Version, wantSettings.Version)
	}

	if recSettings.Width != wantSettings.Width {
		t.Errorf("getAsciicastConfig on file %s found width of %d, want %d.", recPath, recSettings.Width, wantSettings.Width)
	}

	if recSettings.Height != wantSettings.Height {
		t.Errorf("getAsciicastConfig on file %s found height of %d, want %d.", recPath, recSettings.Height, wantSettings.Height)
	}

	if recSettings.Time != wantSettings.Time {
		t.Errorf("getAsciicastConfig on file %s found timestamp of %d, want %d.", recPath, recSettings.Time, wantSettings.Time)
	}

	if recSettings.Env.Shell != wantSettings.Env.Shell {
		// null is clearer when printed than an empty string.
		var got string
		var want string

		if recSettings.Env.Shell == "" {
			got = "null"
		} else {
			got = recSettings.Env.Shell
		}
		if wantSettings.Env.Shell == "" {
			want = "null"
		} else {
			want = wantSettings.Env.Shell
		}
		t.Errorf("getAsciicastConfig on file %s found shell %s, want %s.", recPath, got, want)
	}
	if recSettings.Env.Term != wantSettings.Env.Term {
		// null is clearer when printed than an empty string.
		var got string
		var want string

		if recSettings.Env.Term == "" {
			got = "null"
		} else {
			got = recSettings.Env.Term
		}
		if wantSettings.Env.Term == "" {
			want = "null"
		} else {
			want = wantSettings.Env.Term
		}
		t.Errorf("getAsciicastConfig on file %s found term %s, want %s.", recPath, got, want)
	}
}

// TestGetSceneCastsReturns checks the returned values from
// getSceneCasts. Each value should be a path towards an
// asciicast file. Each path is tested by making sure that
// the file exists and that it has one of the allowed extensions.
func TestGetSceneCastsReturns(t *testing.T) {
	scenePath, err := filepath.Abs("../testdata/scene_1")

	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}

	allFiles, err := ioutil.ReadDir(filepath.Join(scenePath, recordingsPath))

	if err != nil {
		t.Error(err)
	}

	want := 2

	if len(allFiles) != want {
		t.Errorf("testdata in %s should contain %d asciicasts", scenePath, want)
	}

	casts, err := getSceneCasts(scenePath)

	if err != nil {
		t.Errorf("getSceneCasts(%s) returned an error:\n%s", scenePath, err)
	}

	// Checking each file if it is an asciicast.
	for _, file := range casts {
		// File exists
		info, err := os.Stat(file)

		if err != nil {
			t.Errorf("Got error using Stat on one of getSceneCasts returned value:\n%s", err)
		}

		name := info.Name()
		allowedSuffixes := [3]string{"cast", "json", "txt"}
		isOk := false

		for _, suffix := range allowedSuffixes {
			// If file has one of the allowed suffixes, it's valid.
			if strings.HasSuffix(name, suffix) {
				isOk = true
				break
			}
		}

		if !isOk {
			t.Errorf("%s does not have one of the allowed extensions. Allowed extensions are %s", name, allowedSuffixes)
		}
	}
}

// TestGetSceneCastsOnCasts tests getSceneCasts on a scene that
// only contains Asciinema recordings. Testing is done by getting the
// length of the array that is returned by getSceneCasts. Elements
// are not individually checked to make sure that they are really
// paths towards asciicasts.
func TestGetSceneCastsOnCasts(t *testing.T) {
	scenePath, err := filepath.Abs("../testdata/scene_1")
	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}

	allFiles, err := ioutil.ReadDir(filepath.Join(scenePath, recordingsPath))

	if err != nil {
		t.Error(err)
	}

	want := 2

	if len(allFiles) != want {
		t.Errorf("testdata in %s should contain %d asciicasts", scenePath, want)
	}

	casts, err := getSceneCasts(scenePath)

	if err != nil {
		t.Errorf("getSceneCasts(%s) returned an error:\n%s", scenePath, err)
	}

	if len(casts) != want {
		t.Errorf("getSceneCasts(%s) returns an array of len (%d), should be %d", scenePath, len(casts), want)
	}
}

// TestGetSceneCastsWithExtra tests getSceneCasts with extra text files
// that should not be recognized as asciicast files. Those files are
// contained in the second scene of the testdata directory.
func TestGetSceneCastsWithExtra(t *testing.T) {
	scenePath, err := filepath.Abs("../testdata/scene_2")
	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}

	allFiles, err := ioutil.ReadDir(filepath.Join(scenePath, recordingsPath))

	if err != nil {
		t.Error(err)
	}

	filesAmount := 4

	if len(allFiles) != filesAmount {
		t.Errorf("testdata in %s should contain %d asciicasts", scenePath, filesAmount)
	}

	want := 1

	casts, err := getSceneCasts(scenePath)

	if err != nil {
		t.Errorf("getSceneCasts(%s) returned an error:\n%s", scenePath, err)
	}

	if len(casts) != want {
		t.Errorf("getSceneCasts(%s) returns an array of len (%d), should be %d", scenePath, len(casts), want)
	}
}

// TestGetScenePath is checks whether or not the getScenePath function
// returns the proper scene for a file. The file used for this test
// is read_1.txt from testdata/scene_2/read/read_1.txt
func TestGetScenePath(t *testing.T) {
	path, err := filepath.Abs("../testdata/scene_2/read/read_1.txt")
	if err != nil {
		t.Errorf("Error finding testdata: %s", err)
	}
	scenePath, err := getScenePath(path)

	if err != nil {
		t.Error(err)
	}

	splitGot := strings.Split(scenePath, "/")
	got := splitGot[len(splitGot)-1]

	if got != "scene_2" {
		t.Errorf("getScenePath(%s) = %s, want %s", path, got, "scene_2")
	}
}

// TestGetScenePathErr tests getScenePath on a file that does not
// exist to make sure that it returns an error.
func TestGetScenePathErr(t *testing.T) {
	falsePath := "./foobar"
	_, err := os.Stat(falsePath)

	// File should not exist, there should be an error.
	if err == nil {
		t.Errorf("TestGetScenePathErr should be edited. Directory %s should not exist.", falsePath)
	}

	_, err = getScenePath(falsePath)

	if err == nil {
		t.Errorf("getScenePath(%s) should raise an error.", falsePath)
	}
}
