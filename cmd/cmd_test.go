package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestMain is a function that runs before each testing session.
func TestMain(m *testing.M) {
	projectDir := setup()
	defer os.RemoveAll(projectDir) // cleanup
	testsRes := m.Run()
	os.Exit(testsRes)
}

// toCreate contains directories that will be temporarily created
// during each test run.
var toCreate = struct {
	testProject string
	testDir     string
	testFile    string
}{"project", "testDir", "testFile"}

// testData contains paths towards files and directories that will
// be used during testing.
var testData = struct {
	project string
	dir     string
	file    string
	testProject1 string
	noAudio string
}{"", "", "", "", ""}

// setup contains pre test configurations. Everything contained in this
// function will run before each test run. Every error encountered in
// this function exits the program.
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

	// Add other pre test setup here.
	return projectDir
}
