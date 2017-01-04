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

	"holocm.org/cmd/holo/external"
	"holocm.org/lib/holo"
)

var rootDirectory string

func init() {
	rootDirectory = os.Getenv("HOLO_ROOT_DIR")
	if rootDirectory == "" {
		rootDirectory = "/"
	}
}

//RootDirectory returns the environment variable $HOLO_ROOT_DIR, or else the
//default value "/".
func RootDirectory() string {
	return rootDirectory
}

//Configuration contains the parsed contents of /etc/holorc.
type Configuration struct {
	Plugins []holo.Plugin
}

//List config snippets in /etc/holorc.d.
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
		external.Errorf(external.Stderr, err.Error())
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
			pluginID := strings.TrimSpace(strings.TrimPrefix(line, "plugin"))

			var (
				plugin holo.Plugin
				err    error
			)
			if strings.Contains(pluginID, "=") {
				fields := strings.SplitN(pluginID, "=", 2)
				plugin, err = external.NewPluginWithExecutablePath(fields[0], fields[1], NewRuntime(fields[0]))
			} else {
				plugin, err = NewPlugin(pluginID, NewRuntime(pluginID))
			}

			if err == nil {
				result.Plugins = append(result.Plugins, plugin)
			} else {
				if err == external.ErrPluginExecutableMissing {
					//this is not an error because we need a way to uninstall
					//plugins: when the plugin's files are removed, the next
					//"holo apply file:/etc/holorc" will remove them from the
					//holorc, but to be able to run, Holo needs to be able to
					//ignore the missing uninstalled plugin at this point
					external.Warnf(external.Stderr, "Skipping plugin: %s", pluginID)
				} else {
					external.Errorf(external.Stderr, err.Error())
					return nil
				}
			}
		} else {
			//unknown line
			external.Errorf(external.Stderr, "cannot parse configuration: unknown command: %s", line)
			return nil
		}
	}

	//check existence of resource directories
	hasError := false
	for _, plugin := range result.Plugins {
		dir := plugin.(*external.Plugin).Runtime.ResourceDirPath
		fi, err := os.Stat(dir)
		switch {
		case err != nil:
			external.Errorf(external.Stderr, "cannot open %s: %s", dir, err.Error())
			hasError = true
		case !fi.IsDir():
			external.Errorf(external.Stderr, "cannot open %s: not a directory!", dir)
			hasError = true
		}
	}
	if hasError {
		return nil
	}

	//ensure existence of cache and state directories
	for _, plugin := range result.Plugins {
		dirs := []string{plugin.(*external.Plugin).Runtime.CacheDirPath, plugin.(*external.Plugin).Runtime.StateDirPath}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				external.Errorf(external.Stderr, err.Error())
				hasError = true
			}
		}
	}
	if hasError {
		return nil
	}

	return &result
}
