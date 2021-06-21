/*
Copyright © 2021 Etienne Parent <tricky@beon.ca>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func getRecsPaths(projectPath string) []string {
	allPaths := make([]string)
	dirs, err := ioutil.ReadDir(projectPath)
	if err != nil {
		// The function should panic since providing a path
		// that does not exists here is unexpected.
		log.Panic(err)
	}
	for _, dir := range dirs {
		stat, err := os.Stat(dir)
			if err
		if stat.IsDir() && isCast(file) {

		}
	}
}

func isCast(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		log.Panic(err)
	}
	if filepath.Ext(file) == "cast" && !(stat.IsDir()) {
		return true
	}
	return false
}
