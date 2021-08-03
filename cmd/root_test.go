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

// TestIsReadStatementWithRead uses isReadStatement on a project that contains
// audio files. isReadStatement should return true.
func TestIsReadStatementWithRead(t *testing.T) {
	isRead, err := isReadStatement(testData.testProject1)
	if err != nil {
		t.Errorf("isReadStatement(%s) returned error:\n%s", testData.testProject1, err)
	}
	if !isRead {
		t.Errorf("isReadStatement(%s) returned %t, should be %t", testData.testProject1, isRead, !isRead)
	}
}

// TestIsReadStatementWithRead uses isReadStatement on a project that does not
// contain audio files. isReadStatement should return false.
func TestIsReadStatementNoRead(t *testing.T) {
	isRead, err := isReadStatement(testData.noAudio)
	if err != nil {
		t.Errorf("isReadStatement(%s) returned error:\n%s", testData.noAudio, err)
	}
	if !isRead {
		t.Errorf("isReadStatement(%s) returned %t, should be %t", testData.noAudio, isRead, !isRead)
	}
}
