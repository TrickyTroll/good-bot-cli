package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

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

func setup() {
	projectDir, err := ioutil.TempDir("", toCreate.testProject)

	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(projectDir) // clean up

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
	testData.file = filepath.Join(projectDir, file.Name())
}
