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
		t.Errorf(`getDir(%s)`, dir)
	}
}

func TestValidatePath(t *testing.T) {
	var testCases = []struct {
		input string
		want  bool
	}{
		{"", false},
		{"/", true},
		{"/usr/", true},
		{"/foobar/", false},
		{"/usr/local", true},
	}
	for _, test := range testCases {
		if got := validatePath(test.input); got != test.want {
			t.Errorf("validatePath(%v) = %v, want %v", test.input, got, test.want)
		}
	}
}
