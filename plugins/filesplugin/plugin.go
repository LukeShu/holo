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

// Package filesplugin provides a holo.Plugin to provision files.
//
// Directories we use
//
//    target directory      - "{{Runtime.RootDirPath}}"
//    target base directory - "{{Runtime.StateDirPath}}/base"
//    provisioned directory - "{{Runtime.StateDirPath}}/provisioned"
//    resource directory    - "{{Runtime.ResourceDirPath}}"
//
// The flow of information between these looks like (for brevity, this
// graph uses Pacman/libALPM names; see below for how it changes with
// different package managers):
//
//                ,--> resourceDir >----------------------,
//               /                                         \
//    pacman >---                                           \
//               \                                           \
//                '--> targetDir.pacnew >--> targetBaseDir >--`--> provisionedDir
//    meddling                                                           V
//    user >-?-?-?-?-> targetDir                                         |
//                         ^                                             |
//                         |                                             |
//                         `---------------------------------------------'
//
//    (where targetDir.pacnew denotes that the file was placed in
//    targetDir, but with ".pacnew" suffixed to the filename)
//
// Now, we don't just support Pacman, we also support DPKG and RPM.
// Each of these package managers has a way of placing a version of
// the file from the providing package, and a version modified by the
// user (or, in our case: Holo), side-by-side:
//
//    |                         | package manager version | user/holo version |
//    |-------------------------+-------------------------+-------------------|
//    | Pacman case 1           | FILE.pacnew             | FILE              |
//    | Pacman case 2 (< 5.0.0) | FILE                    | FILE.pacorig      |
//    | Pacman case 3           | (deleted)               | FILE.pacsave      |
//    |-------------------------+-------------------------+-------------------|
//    | DPKG case 1             | FILE.dpkg-dist          | FILE              |
//    | DPKG case 1             | FILE                    | FILE.dpkg-old     |
//    |-------------------------+-------------------------+-------------------|
//    | RPM case 1              | FILE.rpmnew             | FILE              |
//    | RPM case 2              | FILE                    | FILE.rpmsave      |
//
// (We don't actually support "Pacman case 2", as it no longer exists
// in modern versions of Pacman; but I figured I'd include it for
// symmetry.)
//
// The details of how the package manager chooses between "case 1" and
// "case 2" are beyond the scope of this documentation; and we don't
// really care.  Whenever we see that the package manager has chosen
// "case 2", we rename the files to the "case 1" names, and proceed as
// if the package manager chose "case 1" (as the "case 1" graph is
// cleaner--the actual file with no suffix only has one "owner").
//
// "case 3" is for when the package has deleted the file, but the user
// (or Holo) has modified the file, and the package manager backs up
// this file to avoid deleting the user's work.  Unfortunately,
// neither DPKG nor RPM offer the user this courtesy; they simply
// delete the file.
//
// ----
//
// We can tell if the user (or another process) has manually changed
// the file from what we provisioned by comparing targetDir with
// provisionedDir.
//
// ----
//
// An "orphan" is a file that was previously provisioned by
// holo-files, but is no longer.  It may or may not still be managed
// by the package manager.  We can enumerate orphans by comparing the
// lists of filenames in targetBaseDir and resourceDir.
//
// When "apply"ing an orphan, there are two courses of action:
// "restore" or "delete".
//
//  - "restore" is for when the file is still managed by the package
//    manager; we restore the package manager version (archived in
//    targetBaseDir).
//  - "delete" is for when the file is no longer managed by the
//    package manager either; we delete the file.
//
// When we "delete", we also delete the .pacsave file.  This might be
// a BUG(lukeshu).
package filesplugin

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/holocm/holo/lib/holo"
)

// FilesPlugin implements holo.Plugin
type FilesPlugin struct {
	Runtime holo.Runtime
}

func (p FilesPlugin) HoloInfo() map[string]string {
	return map[string]string{
		"MIN_API_VERSION": "3",
		"MAX_API_VERSION": "3",
	}
}

func (p FilesPlugin) HoloApply(entityID string, force bool, stdout, stderr io.Writer) holo.ApplyResult {
	e, err := p.getEntity(entityID, stderr)
	if err != nil {
		return holo.NewApplyError(err)
	}
	return e.(*FilesEntity).Apply(force, stdout, stderr)
}

func (p FilesPlugin) HoloDiff(entityID string, stderr io.Writer) (string, string) {
	selectedEntity, err := p.getEntity(entityID, stderr)
	if err != nil {
		return "", ""
	}
	relPath := selectedEntity.(*FilesEntity).relPath
	new := filepath.Join(p.Runtime.StateDirPath+"/provisioned", relPath)
	cur := filepath.Join(p.Runtime.RootDirPath, relPath)
	return new, cur
}

func (p FilesPlugin) getEntity(entityID string, stderr io.Writer) (holo.Entity, error) {
	entities, err := p.HoloScan(stderr)
	if err != nil {
		return nil, err
	}
	for _, entity := range entities {
		if entity.EntityID() == entityID {
			return entity, nil
		}
	}
	fmt.Fprintf(stderr, "!! unknown entity ID \"%s\"\n", entityID)
	return nil, errors.New("")
}

func NewFilesPlugin(r holo.Runtime) holo.Plugin {
	return FilesPlugin{r}
}
