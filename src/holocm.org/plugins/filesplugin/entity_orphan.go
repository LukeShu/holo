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

	"holocm.org/plugins/filesplugin/fileutil"
)

// scanOrphanedTargetBase locates a target file for a given orphaned
// target base and assesses the situation. This logic is grouped in
// one function because it's used by both `holo scan` and `holo
// apply`.
func (target *FilesEntity) scanOrphanedTargetBase() (theTargetPath, strategy, assessment string) {
	targetPath := filepath.Join(target.plugin.Runtime.RootDirPath, target.relPath)
	if fileutil.IsManageableFile(targetPath) {
		return targetPath, "restore", "all repository files were deleted"
	}
	return targetPath, "delete", "target was deleted"
}

// handleOrphanedTargetBase cleans up an orphaned target base.
func (target *FilesEntity) handleOrphanedTargetBase(stdout, stderr io.Writer) []error {
	targetPath, strategy, _ := target.scanOrphanedTargetBase()
	targetBasePath := filepath.Join(target.plugin.Runtime.StateDirPath+"/base", target.relPath)
	lastProvisionedPath := filepath.Join(target.plugin.Runtime.StateDirPath+"/provisioned", target.relPath)

	var errs []error
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	switch strategy {
	case "delete":
		// if the package management left behind additional
		// cleanup targets (most likely a backup of our custom
		// configuration), we can delete these too
		//
		// BUG(lukeshu): We should first compare the .pacsave
		// file to a version in /provisioned to verify that
		// the .pacsave file only has holo-files made changes;
		// otherwise we should leave it behind for the same
		// reason that pacman did.
		cleanupTargets := GetPackageManager(stdout, stderr).AdditionalCleanupTargets(targetPath)
		for _, otherFile := range cleanupTargets {
			fmt.Fprintf(stdout, ">> also deleting %s\n", otherFile)
			appendError(os.Remove(otherFile))
		}
		appendError(os.Remove(lastProvisionedPath))
		appendError(os.Remove(targetBasePath))
	case "restore":
		// target is still there - restore the target base,
		// *but* before that, check if there is an updated
		// target base
		updatedTBPath, reportedTBPath, err := GetPackageManager(stdout, stderr).FindUpdatedTargetBase(targetPath)
		appendError(err)
		if updatedTBPath != "" {
			fmt.Fprintf(stdout, ">> found updated target base: %s -> %s", reportedTBPath, targetPath)
			//use this target base instead of the one in the TargetBaseDirectory
			appendError(os.Remove(targetBasePath))
			targetBasePath = updatedTBPath
		}

		// now really restore the target base
		appendError(os.Remove(lastProvisionedPath))
		appendError(fileutil.MoveFile(targetBasePath, targetPath))
	}

	// TODO(majewsky): cleanup empty directories below
	// StateDirPath+"/provisioned" and StateDirPath+"/base"
	return errs
}
