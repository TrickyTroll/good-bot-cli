package cmd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	testsRes := m.Run()
	os.Exit(testsRes)
}

func TestIsDirectoryOnDir(t *testing.T) {
	isDir := isDirectory(testData.dir)
	if !isDir {
		t.Errorf("isDirectory(%s) = %v", testData.dir, isDir)
	}

}

func TestIsDirectoryOnFile(t *testing.T) {
	isDir := isDirectory(testData.file)
	if isDir {
		t.Errorf("isDirectory(%s) = %v", testData.file, isDir)
	}

}
