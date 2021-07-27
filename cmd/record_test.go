package cmd

import (
	"testing"
)

func TestIsDirectoryOnDir(t *testing.T) {
	isDir, err := isDirectory(testData.dir)
	if err != nil {
		t.Error(err)
	}
	if !isDir {
		t.Errorf("isDirectory(%s) = %v", testData.dir, isDir)
	}
}

func TestIsDirectoryOnFile(t *testing.T) {
	isDir, err := isDirectory(testData.file)
	if err != nil {
		t.Error(err)
	}
	if isDir {
		t.Errorf("isDirectory(%s) = %v", testData.file, isDir)
	}

}
