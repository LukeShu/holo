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

// Command holo is the user interface to Holo.
package entrypoint

import (
	"os"
	"path/filepath"

	"github.com/holocm/holo/cmd/holo/internal/externalplugin"
	"github.com/holocm/holo/cmd/holo/internal/impl"
	"github.com/holocm/holo/lib/holo"
)

//this is populated at compile-time, see Makefile
var version = "unknown"

// Main is the main entry point, but returns the exit code rather than
// calling os.Exit().  This distinction is useful for monobinary and
// testing purposes.
func Main() (exitCode int) {
	rootDir := os.Getenv("HOLO_ROOT_DIR")
	if rootDir == "" {
		rootDir = "/"
	}

	getPlugin := func(id string, arg *string, runtime holo.Runtime) (holo.Plugin, error) {
		if arg == nil {
			_arg := filepath.Join(rootDir, "usr/lib/holo/holo-"+id)
			arg = &_arg
		}
		plugin, err := externalplugin.NewExternalPlugin(id, *arg, runtime)
		if err != nil {
			return nil, err
		}
		return plugin, nil
	}

	return impl.Main(rootDir, version, getPlugin)
}
