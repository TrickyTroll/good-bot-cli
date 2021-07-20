package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var toCreate = struct {
	testProject string
	testDir     string
	testFile    string
}{"project", "testDir", "testFile"}

func setup() {
	projectDir, err := ioutil.TempDir("", toCreate.testProject)

	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(projectDir) // clean up

	ioutil.TempDir(projectDir, toCreate.testDir)
	ioutil.TempFile(projectDir, toCreate.testFile)
}

func TestMain(m *testing.M) {
	setup()
	testsRes := m.Run()
	os.Exit(testsRes)
}
