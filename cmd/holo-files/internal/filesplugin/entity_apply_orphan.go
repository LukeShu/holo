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

package filesplugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
)

// scanOrphan locates an entity for a given orphaned entity and
// assesses the situation. This logic is grouped in one function
// because it's used by both `holo scan` and `holo apply`.
func (entity *FilesEntity) scanOrphan() (targetPath, strategy, assessment string) {
	targetPath = filepath.Join(entity.plugin.Runtime.RootDirPath, entity.relPath)
	if fileutil.IsManageableFile(targetPath) {
		return targetPath, "restore", "all repository files were deleted"
	}
	return targetPath, "delete", "target was deleted"
}

// applyOrphan cleans up an orphaned entity.
func (entity *FilesEntity) applyOrphan(stdout, stderr io.Writer) []error {
	_, strategy, _ := entity.scanOrphan()
	basePath := filepath.Join(entity.plugin.Runtime.StateDirPath, "base", entity.relPath)

	var errs []error
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	current, err := entity.GetCurrent()
	if !os.IsNotExist(err) {
		appendError(err)
	}

	provisioned, err := entity.GetProvisioned()
	appendError(err)

	switch strategy {
	case "delete":
		//if the package management left behind additional cleanup targets
		//(most likely a backup of our custom configuration), we can delete
		//these too
		cleanupTargets := GetPackageManager(entity.plugin.Runtime.RootDirPath, stdout, stderr).AdditionalCleanupTargets(current.Path)
		for _, path := range cleanupTargets {
			otherFile, err := fileutil.NewFileBuffer(path)
			if err != nil {
				continue
			}
			if otherFile.EqualTo(provisioned) {
				fmt.Fprintf(stdout, ">> also deleting %s\n", otherFile.Path)
				appendError(os.Remove(otherFile.Path))
			}
		}

		appendError(os.Remove(provisioned.Path))
		appendError(os.Remove(basePath))
	case "restore":
		//target is still there - restore the target base, *but* before that,
		//check if there is an updated target base
		updatedTBPath, reportedTBPath, err := GetPackageManager(entity.plugin.Runtime.RootDirPath, stdout, stderr).FindUpdatedTargetBase(current.Path)
		appendError(err)
		if updatedTBPath != "" {
			fmt.Fprintf(stdout, ">> found updated target base: %s -> %s", reportedTBPath, current.Path)
			//use this target base instead of the one in the BaseDirectory
			appendError(os.Remove(basePath))
			basePath = updatedTBPath
		}

		appendError(os.Remove(provisioned.Path))
		appendError(fileutil.MoveFile(basePath, current.Path))
	}

	// TODO(majewsky): cleanup empty directories below
	// StateDirPath+"/base" and StateDirPath+"/provisioned"
	return errs
}
