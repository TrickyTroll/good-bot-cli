package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceParams tests the returned value by replaceParam.
// The params are unmarshalled and tested before and after passing
// through replaceParam.
//
// This function uses data from the croptests directory in testdata.
func TestReplaceParams(t *testing.T) {

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

	casts := getSceneCasts(scenePath)

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

	casts := getSceneCasts(scenePath)

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

	casts := getSceneCasts(scenePath)

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

	casts := getSceneCasts(scenePath)

	if len(casts) != want {
		t.Errorf("getSceneCasts(%s) returns an array of len (%d), should be %d", scenePath, len(casts), want)
	}
}

// TestGetScenePath is checks wether or not the getScenePath function
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
