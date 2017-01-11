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
	"io/ioutil"
	"os"
	"path/filepath"

	"holocm.org/cmd/holo/impl"
	"holocm.org/cmd/holo/output"
	"holocm.org/lib/holo"
)

type RuntimeManager struct {
	rootDir  string
	cacheDir string
}

func NewRuntimeManager(rootDir string) (*RuntimeManager, error) {
	// TODO(lukeshu): Consider inspecting os.TempDir() to see if
	// it is below rootDir.  I don't think it's important to do so
	// because ioutil.TempDir() avoids conflicts.
	cacheDir, err := ioutil.TempDir(os.TempDir(), "holo.")
	if err != nil {
		return nil, err
	}
	return &RuntimeManager{rootDir: rootDir, cacheDir: cacheDir}, nil
}

func (r *RuntimeManager) Close() {
	_ = os.RemoveAll(r.cacheDir) // fail silently
}

func (r *RuntimeManager) NewRuntime(id string) holo.Runtime {
	return holo.Runtime{
		APIVersion:      3,
		RootDirPath:     r.rootDir,
		ResourceDirPath: filepath.Join(r.rootDir, "usr/share/holo", id),
		CacheDirPath:    filepath.Join(r.cacheDir, id),
		StateDirPath:    filepath.Join(r.rootDir, "var/lib/holo", id),
	}
}

func (r *RuntimeManager) GetPlugins(config []impl.PluginConfig) []*impl.PluginHandle {
	plugins := []*impl.PluginHandle{} // non nil
	for _, pluginConfig := range config {
		pluginHandle, err := impl.NewPluginHandle(
			pluginConfig.ID,
			pluginConfig.Arg,
			r.NewRuntime(pluginConfig.ID),
			GetPlugin)
		if err != nil {
			if os.IsNotExist(err) {
				// this is not an error because we need a way
				// to uninstall plugins: when the plugin's
				// files are removed, the next "holo apply
				// file:/etc/holorc" will remove them from the
				// holorc, but to be able to run, Holo needs
				// to be able to ignore the missing
				// uninstalled plugin at this point
				output.Warnf(output.Stderr, "%s", err.Error())
				output.Warnf(output.Stderr, "Skipping plugin: %s", pluginConfig)
				continue
			} else {
				output.Errorf(output.Stderr, "%s", err.Error())
				return nil
			}
		}
		plugins = append(plugins, pluginHandle)
	}

	return plugins
}
