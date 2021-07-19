package cmd

import (
	"strings"
	"testing"
)

func TestGetDir(t *testing.T) {
	path := "./cmd/root.go"
	dir := getDir(path)
	splitDir := strings.Split(dir, "/")
	stem := splitDir[len(splitDir)-1]
	if !(stem == "cmd") {
		t.Error(`getDir(%s)`, dir)
	}
}
