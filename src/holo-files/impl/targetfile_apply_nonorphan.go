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
func (target *TargetFile) apply(haveForce bool) (skipReport bool, errs []error) {
	// determine the related paths
	targetPath := target.PathIn(common.TargetDirectory())
	targetBasePath := target.PathIn(common.TargetBaseDirectory())
	lastProvisionedPath := target.PathIn(common.ProvisionedDirectory())

	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// step 1: Check if a system update installed a new version of
	// the stock configuration
	//
	// This has to come first because it might shuffle some files
	// around, and if we do anything else first, we might end up
	// stat()ing the wrong file.
	updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(targetPath)
	if err != nil {
		return false, []error{err}
	}
	if updatedTBPath != "" {
		//an updated stock configuration is available at updatedTBPath
		fmt.Printf(">> found updated target base: %s -> %s\n", reportedTBPath, targetBasePath)
		err := common.CopyFile(updatedTBPath, targetBasePath)
		if err != nil {
			return false, []error{fmt.Errorf("Cannot copy %s to %s: %s", updatedTBPath, targetBasePath, err.Error())}
		}
		_ = os.Remove(updatedTBPath) //this can fail silently
	}

	// step 2: Load the 3 versions into memory.
	targetBuffer, err := common.NewFileBuffer(targetPath)
	if !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			pe.Op = "skipping"
			pe.Path = "target"
		}
		appendError(err)
	}

	baseBuffer, err := common.NewFileBuffer(targetBasePath)
	if !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			pe.Op = "skipping"
			pe.Path = "target"
		}
		appendError(err)
	}

	lastProvisionedBuffer, err := common.NewFileBuffer(lastProvisionedPath)
	if !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			pe.Op = "skipping"
			pe.Path = "target"
		}
		appendError(err)
	}

	if len(errs) > 0 {
		return false, errs
	}

	var strategy func(base, provisioned, target common.FileBuffer, haveForce bool) (skipReport bool, errs []error)
	if len(target.repoEntries) == 0 {
		if targetBuffer.Manageable {
			strategy = target.applyRestore
		} else {
			strategy = target.applyDelete
		}
	} else {
		strategy = target.applyProvision
	}
	//TODO: cleanup empty directories below TargetBaseDirectory() and ProvisionedDirectory()
	return strategy(baseBuffer, lastProvisionedBuffer, targetBuffer, haveForce)
}

func (tf *TargetFile) applyProvision(base, provisioned, target common.FileBuffer, haveForce bool) (skipReport bool, errs []error) {
	var err error
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// step 1: figure out if we need to shuffle files around
	if target.Manageable && !base.Manageable {
		// if we don't have a target base yet, the file at
		// targetPath *is* the targetBase; which we have to
		// copy now
		targetBaseDir := filepath.Dir(base.Path)
		err = os.MkdirAll(targetBaseDir, 0755)
		if err != nil {
			appendError(fmt.Errorf("Cannot create directory %s: %s", targetBaseDir, err.Error()))
			return
		}

		err = target.Write(base.Path)
		if err != nil {
			appendError(fmt.Errorf("Cannot copy %s to %s: %s", target.Path, base.Path, err.Error()))
			return
		}
		tmp := target
		tmp.Path = base.Path
		base = tmp
	}

	// step 2: apply the repo files (in memory)

	if !base.Manageable {
		appendError(&os.PathError{
			Op:   "skipping",
			Path: "target",
			Err:  common.ErrNotManageable,
		})
		return
	}

	buffer, err := tf.Render(base)
	if err != nil {
		appendError(err)
		return
	}

	// step 3: save to the filesystem
	//
	// We need to save `buffer` to 2 different files:
	//  - the last-provisioned file
	//  - the actual file

	// actual file
	if !target.Manageable || !buffer.Equal(target) {
		if !haveForce {
			if !target.Manageable {
				appendError(ErrNeedForceToRestore)
				return
			}
			if provisioned.Manageable && !buffer.Equal(provisioned) {
				appendError(ErrNeedForceToOverwrite)
				return
			}
		}
		// Do a $target.holonew -> $target shuffle so that
		// $target is updated atomically (to ensure that there
		// is always a valid file at $target)
		err = buffer.Write(target.Path + ".holonew")
		if err != nil {
			appendError(err)
			return
		}
		err = os.Rename(target.Path+".holonew", target.Path)
		if err != nil {
			appendError(err)
			return
		}
	} else {
		skipReport = true
	}

	// last-provisioned file
	if !provisioned.Manageable || !buffer.Equal(provisioned) {
		provisionedDir := filepath.Dir(provisioned.Path)
		err = os.MkdirAll(provisionedDir, 0755)
		if err == nil {
			err = buffer.Write(provisioned.Path)
		}
		appendError(err)
	}

	return
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
