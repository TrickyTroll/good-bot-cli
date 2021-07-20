package cmd

import (
	"io/ioutil"
	"os"
)

var toCreate = struct {
	testProject string
	testDir     string
	testFile    string
}{"project", "testDir", "testFile"}

func makeTempProject() (string, error) {
	projectDir, err := ioutil.TempDir("", toCreate.testProject)

	if err != nil {
		return "", err
	}
	defer os.RemoveAll(projectDir) // clean up

	ioutil.TempDir(projectDir, toCreate.testDir)
	ioutil.TempFile(projectDir, toCreate.testFile)

	return projectDir, nil
}

var tempProject, err = makeTempProject()
