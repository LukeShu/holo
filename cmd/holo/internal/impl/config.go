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

package impl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/holocm/holo/cmd/holo/internal/externalplugin"
	"github.com/holocm/holo/cmd/holo/internal/output"
	"github.com/holocm/holo/lib/holo"
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

func GetPlugin(id string, arg *string, runtime holo.Runtime) (holo.Plugin, error) {
	if arg == nil {
		_arg := filepath.Join(RootDirectory(), "usr/lib/holo/holo-"+id)
		arg = &_arg
	}
	plugin, err := externalplugin.NewExternalPlugin(id, *arg, runtime)
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

func NewRuntime(id string) holo.Runtime {
	return holo.Runtime{
		APIVersion:      3,
		RootDirPath:     RootDirectory(),
		ResourceDirPath: filepath.Join(RootDirectory(), "usr/share/holo/"+id),
		CacheDirPath:    filepath.Join(CachePath(), id),
		StateDirPath:    filepath.Join(RootDirectory(), "var/lib/holo/"+id),
	}
}

//Configuration contains the parsed contents of /etc/holorc.
type Configuration struct {
	Plugins []*PluginHandle
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
			pluginSpec := strings.TrimSpace(strings.TrimPrefix(line, "plugin"))

			var (
				pluginID  string
				pluginArg *string
			)
			if strings.Contains(pluginSpec, "=") {
				fields := strings.SplitN(pluginSpec, "=", 2)
				pluginID = fields[0]
				pluginArg = &fields[1]
			} else {
				pluginID = pluginSpec
			}
			plugin, err := NewPluginHandle(pluginID, pluginArg, NewRuntime(pluginID), GetPlugin)

			if err == nil {
				result.Plugins = append(result.Plugins, plugin)
			} else {
				if os.IsNotExist(err) {
					//this is not an error because we need a way to uninstall
					//plugins: when the plugin's files are removed, the next
					//"holo apply file:/etc/holorc" will remove them from the
					//holorc, but to be able to run, Holo needs to be able to
					//ignore the missing uninstalled plugin at this point
					output.Warnf(output.Stderr, "%s", err.Error())
					output.Warnf(output.Stderr, "Skipping plugin: %s", pluginSpec)
				} else {
					output.Errorf(output.Stderr, "%s", err.Error())
					return nil
				}
			}
		} else {
			//unknown line
			output.Errorf(output.Stderr, "cannot parse configuration: unknown command: %s", line)
			return nil
		}
	}

	//check existence of resource directories
	hasError := false
	for _, plugin := range result.Plugins {
		dir := plugin.Runtime.ResourceDirPath
		fi, err := os.Stat(dir)
		switch {
		case err != nil:
			output.Errorf(output.Stderr, "cannot open %s: %s", dir, err.Error())
			hasError = true
		case !fi.IsDir():
			output.Errorf(output.Stderr, "cannot open %s: not a directory!", dir)
			hasError = true
		}
	}
	if hasError {
		return nil
	}

	//ensure existence of cache and state directories
	for _, plugin := range result.Plugins {
		dirs := []string{plugin.Runtime.CacheDirPath, plugin.Runtime.StateDirPath}
		for _, dir := range dirs {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				output.Errorf(output.Stderr, err.Error())
				hasError = true
			}
		}
	}
	if hasError {
		return nil
	}

	return &result
}
