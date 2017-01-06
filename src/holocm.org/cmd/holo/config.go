/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"holocm.org/cmd/holo/impl"
	"holocm.org/cmd/holo/output"
)

//Configuration contains the parsed contents of /etc/holorc.
type Configuration struct {
	Plugins []*impl.PluginHandle
}

// List config snippets in /etc/holorc.d.
func listConfigSnippets() ([]string, error) {
	dirPath := filepath.Join(RootDirectory(), "etc/holorc.d")
	dir, err := os.Open(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			//non-existence of the directory is acceptable
			return nil, nil
		}
		return nil, err
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, fi := range fis {
		if !fi.Mode().IsDir() {
			paths = append(paths, filepath.Join(dirPath, fi.Name()))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

//The part of ReadConfiguration that reads all the holorc files.
func readConfigLines() ([]string, error) {
	//enumerate snippets
	paths, err := listConfigSnippets()
	if err != nil {
		return nil, err
	}
	//holorc is read at the very end, after all snippets
	paths = append(paths, filepath.Join(RootDirectory(), "etc/holorc"))

	//read snippets in order
	var lines []string
	for _, path := range paths {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %s", path, err.Error())
		}
		lines = append(lines, strings.SplitN(strings.TrimSpace(string(contents)), "\n", -1)...)
	}
	return lines, nil
}

//ReadConfiguration reads the configuration file /etc/holorc.
func ReadConfiguration() *Configuration {
	lines, err := readConfigLines()
	if err != nil {
		output.Errorf(output.Stderr, err.Error())
		return nil
	}

	var result Configuration
	for _, line := range lines {
		//ignore comments and empty lines
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		//collect plugin IDs
		if strings.HasPrefix(line, "plugin ") {
			plugin, err := getPlugin2(strings.TrimSpace(strings.TrimPrefix(line, "plugin")))
			if err != nil {
				output.Errorf(output.Stderr, "%s", err.Error())
				return nil
			}
			if plugin != nil {
				result.Plugins = append(result.Plugins, plugin)
			}
		} else {
			//unknown line
			output.Errorf(output.Stderr, "cannot parse configuration: unknown command: %s", line)
			return nil
		}
	}

	if !Setup(result.Plugins) {
		return nil
	}
	return &result
}
