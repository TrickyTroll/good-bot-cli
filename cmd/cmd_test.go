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

var TestProject1 string = "../testdata/project_1"

var toCreate = struct {
	testProject string
	testDir     string
	testFile    string
}{"project", "testDir", "testFile"}

var testData = struct {
	project string
	dir     string
	file    string
}{"", "", ""}

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

	return projectDir
}
