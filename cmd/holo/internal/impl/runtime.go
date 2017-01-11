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
