package cmd

import (
	"testing"
)

func TestIsDirectoryOnDir(t *testing.T) {
	isDir := isDirectory(toCreate.testDir)
	if !isDir {
		t.Errorf("isDirectory(%s) = %v", toCreate.testDir, isDir)
	}

}

func TestIsDirectoryOnFile(t *testing.T) {
	isDir := isDirectory(toCreate.testFile)
	if isDir {
		t.Errorf("isDirectory(%s) = %v", toCreate.testDir, isDir)
	}

}
