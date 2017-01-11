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
	"os"
	"path/filepath"

	"holocm.org/cmd/holo/impl"
	"holocm.org/cmd/holo/output"
	"holocm.org/lib/holo"
	"holocm.org/plugins/externalplugin"
)

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

func GetPlugins(config []PluginConfig) []*impl.PluginHandle {
	plugins := []*impl.PluginHandle{} // non nil
	for _, pluginConfig := range config {
		pluginHandle, err := impl.NewPluginHandle(
			pluginConfig.ID,
			pluginConfig.Arg,
			NewRuntime(pluginConfig.ID),
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

	hasError := false
	for _, plugin := range plugins {
		if !SetupRuntime(plugin.Runtime) {
			hasError = true
		}
	}
	if hasError {
		return nil
	}

	return plugins
}
