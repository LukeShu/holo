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

	"holocm.org/cmd/holo/output"
	"holocm.org/lib/holo"
)

var (
	rootDir  string
	cacheDir string
)

func init() {
	rootDir = os.Getenv("HOLO_ROOT_DIR")
	if rootDir == "" {
		rootDir = "/"
	}

	// BUG(lukeshu): Consider inspecting os.TempDir() to see if it
	// is below rootDir.  I don't think it's important to do so
	// because us ioutil.TempDir() avoids conflicts.
	var err error
	cacheDir, err = ioutil.TempDir(os.TempDir(), "holo.")
	if err != nil {
		output.Errorf(output.Stderr, err.Error())
		os.Exit(255)
	}
}

// RootDirectory returns the environment variable $HOLO_ROOT_DIR, or
// else the default value "/".
func RootDirectory() string {
	return rootDir
}

func Exit(code int) {
	_ = os.RemoveAll(cacheDir) // fail silently
	os.Exit(code)
}

func NewRuntime(id string) holo.Runtime {
	return holo.Runtime{
		APIVersion:      3,
		RootDirPath:     RootDirectory(),
		ResourceDirPath: filepath.Join(RootDirectory(), "usr/share/holo/"+id),
		CacheDirPath:    filepath.Join(cacheDir, id),
		StateDirPath:    filepath.Join(RootDirectory(), "var/lib/holo/"+id),
	}
}
