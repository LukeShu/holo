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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"holocm.org/lib/holo"
	"holocm.org/plugins/filesplugin/fileutil"
)

// Apply performs the complete application algorithm for the given
// FilesEntity.  This includes taking a copy of the target base if
// necessary, applying all repository entries, and saving the result
// in the target path with the correct file metadata.
func (target *FilesEntity) apply(haveForce bool, stdout, stderr io.Writer) (holo.ApplyResult, error) {
	// determine the related paths
	targetPath := filepath.Join(target.plugin.Runtime.RootDirPath, target.relPath)
	targetBasePath := filepath.Join(target.plugin.Runtime.StateDirPath+"/base", target.relPath)

	// step 1: check if a system update installed a new version of
	// the stock configuration
	//
	// This has to come before reading the targetPath because the
	// package manager might shuffle around files to simplify
	// things.
	updatedTBPath, reportedTBPath, err := GetPackageManager(stdout, stderr).FindUpdatedTargetBase(targetPath)
	if err != nil {
		return nil, err
	}
	if updatedTBPath != "" {
		// an updated stock configuration is available at
		// updatedTBPath (but show it to the user as
		// reportedTBPath).
		fmt.Fprintf(stdout, ">> found updated target base: %s -> %s", reportedTBPath, targetBasePath)
		err := fileutil.MoveFile(updatedTBPath, targetBasePath) // TODO(lukeshu): os.Rename()?
		if err != nil {
			return nil, fmt.Errorf("Cannot move %s to %s: %s", updatedTBPath, targetBasePath, err.Error())
		}
	}

	// step 2: Read the target file in.
	//
	// - option 1: there is a manageable file in the target
	//   location (this target file is either the target base from
	//   the application package or the product of a previous
	//   Apply run)
	//
	// - option 2: the target file was deleted, but we have a
	//   target base that we can start from
	forceReprovision := false
	targetBuffer, err := fileutil.NewFileBuffer(targetPath, false)
	if err != nil {
		targetBuffer, err = fileutil.NewFileBuffer(targetBasePath, false)
		targetBuffer.Path = targetPath
		if err != nil {
			return nil, errors.New("skipping target: not a manageable file")
		}
		if !haveForce {
			return holo.ApplyExternallyDeleted, nil
		}
		forceReprovision = true
	}

	// step 3: verify that the version at targetPath has not been
	// tampered with.
	var lastProvisionedBuffer fileutil.FileBuffer
	lastProvisionedPath := filepath.Join(target.plugin.Runtime.StateDirPath+"/provisioned", target.relPath)
	lastProvisionedBuffer, err = fileutil.NewFileBuffer(lastProvisionedPath, false)
	lastProvisionedBuffer.Path = targetPath
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if !os.IsNotExist(err) {
		if targetBuffer != lastProvisionedBuffer {
			if !haveForce {
				return holo.ApplyExternallyChanged, nil
			}
			forceReprovision = true
		}
	}

	// step 4: apply the repo files

	// Load the target base into a buffer as the start for the
	// application algorithm.
	buffer, err := fileutil.NewFileBuffer(targetBasePath, false)
	if os.IsNotExist(err) {
		// if we don't have a target base yet, the file at
		// targetPath *is* the targetBase which we have to
		// copy now
		targetBaseDir := filepath.Dir(targetBasePath)
		err = os.MkdirAll(targetBaseDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Cannot create directory %s: %s", targetBaseDir, err.Error())
		}
		err = targetBuffer.Write(targetBasePath)
		if err != nil {
			return nil, fmt.Errorf("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
		}
		buffer, err = fileutil.NewFileBuffer(targetBasePath, false)
	}
	if err != nil {
		return nil, err
	}
	buffer.Path = targetPath

	// Get the list of RepoFiles to apply to buffer.
	repoEntries := target.RepoEntries()

	// Optimization: check if we can skip any steps.
	firstStep := 0
	for idx, repoFile := range repoEntries {
		if repoFile.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}
	repoEntries = repoEntries[firstStep:]

	// Apply all the applicable repo files in order
	for _, repoFile := range repoEntries {
		buffer, err = repoFile.ApplyTo(buffer, stdout, stderr)
		if err != nil {
			return nil, err
		}
	}

	// step 5: write the results to the filesystem.

	// Don't actually hit the filesystem if nothing has changed.
	if buffer == lastProvisionedBuffer && !forceReprovision {
		// since we did not do anything, don't report this
		return holo.ApplyAlreadyApplied, nil
	}

	// save a copy of the provisioned config file to check for
	// manual modifications in the next Apply() run
	provisionedDir := filepath.Dir(lastProvisionedPath)
	err = os.MkdirAll(provisionedDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("Cannot create directory %s: %s", provisionedDir, err.Error())
	}
	err = buffer.Write(lastProvisionedPath)
	if err != nil {
		return nil, fmt.Errorf("Cannot write %s: %s", lastProvisionedPath, err.Error())
	}

	// write the result buffer to the target location and copy
	// owners/permissions from target base to target file
	newTargetPath := targetPath + ".holonew"
	err = buffer.Write(newTargetPath)
	if err != nil {
		return nil, err
	}
	// move $target.holonew -> $target atomically (to ensure that
	// there is always a valid file at $target)
	err = os.Rename(newTargetPath, targetPath)
	if err != nil {
		return nil, err
	}
	return holo.ApplyApplied, nil
}
