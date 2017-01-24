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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"../common"
	"../platform"
)

//applyNonOrphan performs the complete application algorithm for the given TargetFile.
//This includes taking a copy of the target base if necessary, applying all
//repository entries, and saving the result in the target path with the correct
//file metadata.
func (target *TargetFile) applyNonOrphan(haveForce bool) (skipReport bool, err error) {
	//determine the related paths
	targetPath := target.PathIn(common.TargetDirectory())
	targetBasePath := target.PathIn(common.TargetBaseDirectory())

	// step 1: Check if a system update installed a new version of
	// the stock configuration
	//
	// This has to come first because it might shuffle some files
	// around, and if we do anything else first, we might end up
	// stat()ing the wrong file.
	updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(targetPath)
	if err != nil {
		return false, err
	}
	if updatedTBPath != "" {
		//an updated stock configuration is available at updatedTBPath
		fmt.Printf(">> found updated target base: %s -> %s", reportedTBPath, targetBasePath)
		err := common.CopyFile(updatedTBPath, targetBasePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", updatedTBPath, targetBasePath, err.Error())
		}
		_ = os.Remove(updatedTBPath) //this can fail silently
	}

	// step 2: Load the current version into memory.
	needForcefulReprovision := false
	targetBuffer, err := common.NewFileBuffer(targetPath)
	if os.IsNotExist(err) {
		targetBuffer, err = common.NewFileBuffer(targetBasePath)
		targetBuffer.Path = targetPath
		if err != nil {
			return false, errors.New("skipping target: not a manageable file")
		}
		if !haveForce {
			return false, ErrNeedForceToRestore
		}
		needForcefulReprovision = true
	}
	if err != nil {
		return false, errors.New("skipping target: not a manageable file")
	}

	// step 3: Load the base version into memory.
	baseBuffer, err := common.NewFileBuffer(targetBasePath)
	baseBuffer.Path = targetPath
	if os.IsNotExist(err) {
		// if we don't have a target base yet, the file at
		// targetPath *is* the targetBase; which we have to
		// copy now
		targetBaseDir := filepath.Dir(targetBasePath)
		err = os.MkdirAll(targetBaseDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot create directory %s: %s", targetBaseDir, err.Error())
		}

		err = targetBuffer.Write(targetBasePath)
		if err != nil {
			return false, fmt.Errorf("Cannot copy %s to %s: %s", targetPath, targetBasePath, err.Error())
		}
		baseBuffer = targetBuffer
	}
	if err != nil {
		return false, err
	}

	// step 4: apply the repo files (in memory)

	buffer, err := target.Render(baseBuffer)
	if err != nil {
		return false, err
	}

	// step 5: Load the last-provisioned version into memory.
	lastProvisionedPath := target.PathIn(common.ProvisionedDirectory())
	lastProvisionedBuffer, err := common.NewFileBuffer(lastProvisionedPath)
	lastProvisionedBuffer.Path = targetPath
	haveLastProvisionedBuffer := !os.IsNotExist(err)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	// step 6: save to the filesystem
	//
	// we have 2 files to hit:
	//  - the last-provisioned file
	//  - the actual file

	// last-provisioned file
	if buffer != lastProvisionedBuffer || !haveLastProvisionedBuffer {
		provisionedDir := filepath.Dir(lastProvisionedPath)
		err = os.MkdirAll(provisionedDir, 0755)
		if err != nil {
			return false, fmt.Errorf("Cannot write %s: %s", lastProvisionedPath, err.Error())
		}
		err = buffer.Write(lastProvisionedPath)
		if err != nil {
			return false, err
		}
	}

	// actual file
	if buffer != targetBuffer || needForcefulReprovision {
		if haveLastProvisionedBuffer && targetBuffer != lastProvisionedBuffer {
			if !haveForce {
				return false, ErrNeedForceToOverwrite
			}
		}
		// Do the $target.holonew -> $target shuffle so that
		// $target is updated atomically (to ensure that there
		// is always a valid file at $target)
		err = buffer.Write(targetPath + ".holonew")
		if err != nil {
			return false, err
		}
		err = os.Rename(targetPath+".holonew", targetPath)
		if err != nil {
			return false, err
		}
		return false, nil
	} else {
		return true, nil
	}
}

//Render applies all the repo files for this TargetFile onto the target base.
func (t *TargetFile) Render(baseBuffer common.FileBuffer) (common.FileBuffer, error) {
	// Optimization: check if we can skip any application steps
	firstStep := 0
	repoEntries := t.RepoEntries()
	for idx, repoFile := range repoEntries {
		if repoFile.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}
	repoEntries = repoEntries[firstStep:]

	//apply all the applicable repo files in order
	buffer := baseBuffer
	var err error
	for _, repoFile := range repoEntries {
		buffer, err = repoFile.ApplyTo(buffer)
		if err != nil {
			return common.FileBuffer{}, err
		}
	}

	return buffer, nil
}
