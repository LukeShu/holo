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
	"path/filepath"

	"holocm.org/cmd/holo/external"
	"holocm.org/lib/holo"
)

//NewPlugin creates a new Plugin.
func NewPlugin(id string, r holo.Runtime) (holo.Plugin, error) {
	executablePath := filepath.Join(RootDirectory(), "usr/lib/holo/holo-"+id)
	p, err := external.NewPluginWithExecutablePath(id, executablePath, r)
	return p, err
}

func NewRuntime(id string) holo.Runtime {
	return holo.Runtime{
		RootDirPath: RootDirectory(),

		ResourceDirPath: filepath.Join(RootDirectory(), "usr/share/holo/"+id),
		CacheDirPath:    filepath.Join(CachePath(), id),
		StateDirPath:    filepath.Join(RootDirectory(), "var/lib/holo/"+id),
	}
}
