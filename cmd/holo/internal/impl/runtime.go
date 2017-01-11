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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/holocm/holo/cmd/holo/internal/output"
	"github.com/holocm/holo/lib/holo"
)

var (
	rootDir   string
	cachePath string
)

//WithCacheDirectory executes the worker function after having set up a cache
//directory, and ensures that the cache directory is cleaned up afterwards.
func WithCacheDirectory(worker func() (exitCode int)) (exitCode int) {
	var err error
	cachePath, err = ioutil.TempDir(os.TempDir(), "holo.")
	if err != nil {
		output.Errorf(output.Stderr, err.Error())
		return 255
	}

	//ensure that the cache is removed even if worker() panics
	defer func() {
		_ = os.RemoveAll(cachePath) //failure to cleanup is non-fatal
		cachePath = ""
	}()

	return worker()
}

//CachePath returns the path below which plugin cache directories can be allocated.
func CachePath() string {
	if cachePath == "" {
		panic("Tried to use cachePath outside WithCacheDirectory() call!")
	}
	return cachePath
}

func init() {
	rootDir = os.Getenv("HOLO_ROOT_DIR")
	if rootDir == "" {
		rootDir = "/"
	}
}

// RootDirectory returns the environment variable $HOLO_ROOT_DIR, or
// else the default value "/".
func RootDirectory() string {
	return rootDir
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

func SetupRuntime(r holo.Runtime) bool {
	hasError := false
	if _, err := os.Stat(r.ResourceDirPath + "/"); err != nil {
		output.Errorf(output.Stderr, "Resource directory cannot be opened: %q: %v", r.ResourceDirPath, err)
		hasError = true
	}
	if err := os.MkdirAll(r.StateDirPath, 0755); err != nil {
		output.Errorf(output.Stderr, "State directory cannot be created: %q: %v", r.StateDirPath, err)
		hasError = true
	}
	if err := os.MkdirAll(r.CacheDirPath, 0755); err != nil {
		output.Errorf(output.Stderr, "Cache directory cannot be created: %q: %v", r.CacheDirPath, err)
		hasError = true
	}
	return !hasError
}
