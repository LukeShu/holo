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
	"strconv"

	"holocm.org/cmd/holo/output"
	"holocm.org/lib/holo"
)

type PluginHandle struct {
	ID      string
	Plugin  holo.Plugin
	Runtime holo.Runtime
	Info    map[string]string
}

type PluginGetter func(id string, arg *string, runtime holo.Runtime) (holo.Plugin, error)

func NewPluginHandle(id string, arg *string, runtime holo.Runtime, getPlugin PluginGetter) (*PluginHandle, error) {
	plugin, err := getPlugin(id, arg, runtime)
	handle := &PluginHandle{
		ID:      id,
		Plugin:  plugin,
		Runtime: runtime,
		Info:    nil,
	}

	// grab/cache metadata
	handle.Info = handle.Plugin.HoloInfo()
	if handle.Info == nil {
		return nil, fmt.Errorf("plugin holo-%s: \"info\" operation failed", handle.ID)
	}
	err = checkVersion(handle, runtime.APIVersion)
	if err != nil {
		return nil, err
	}

	return handle, nil
}

func checkVersion(handle *PluginHandle, version int) error {
	minVersion, err := strconv.Atoi(handle.Info["MIN_API_VERSION"])
	if err != nil {
		return err
	}
	maxVersion, err := strconv.Atoi(handle.Info["MAX_API_VERSION"])
	if err != nil {
		return err
	}
	if minVersion > version || maxVersion < version {
		return fmt.Errorf(
			"plugin holo-%s is incompatible with this Holo (plugin min: %d, plugin max: %d, Holo: %d)",
			handle.ID, minVersion, maxVersion, version,
		)
	}
	return nil
}

func (handle *PluginHandle) Scan() ([]*EntityHandle, error) {
	entities, err := handle.Plugin.HoloScan(output.Stderr)
	if err != nil {
		return nil, err
	}
	ret := make([]*EntityHandle, len(entities))
	for i := range entities {
		ret[i] = &EntityHandle{
			PluginHandle: handle,
			Entity:       entities[i],
		}
	}
	return ret, nil
}
