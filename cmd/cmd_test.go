package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	projectDir := setup()
	defer os.RemoveAll(projectDir) // cleanup
	testsRes := m.Run()
	os.Exit(testsRes)
}

var toCreate = struct {
	testProject string
	testDir     string
	testFile    string
}{"project", "testDir", "testFile"}

var testData = struct {
	project string
	dir     string
	file    string
	testProject1 string
	noAudio string
}{"", "", "", "", ""}

func setup() string {
	projectDir, err := ioutil.TempDir("", toCreate.testProject)

	if err != nil {
		log.Fatal(err)
	}

	dirname, err := ioutil.TempDir(projectDir, toCreate.testDir)
	if err != nil {
		log.Fatal(err)
	}

	file, err := ioutil.TempFile(projectDir, toCreate.testFile)
	if err != nil {
		log.Fatal(err)
	}

	file.Write([]byte("Hello, world!"))

	testData.project = projectDir
	testData.dir = dirname
	testData.file = filepath.Join(file.Name())

	testProject1, err := filepath.Abs("../testdata/project_1")

	if err != nil {
		log.Fatal(err)
	}

	testNoAudio, err := filepath.Abs("../testdata/no_audio")

	if err != nil {
		log.Fatal(err)
	}

	testData.testProject1 = testProject1
	testData.noAudio = testNoAudio

	return projectDir
}
