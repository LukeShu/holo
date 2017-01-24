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

	"../common"
	"../platform"
)

func (tf *TargetFile) applyDelete(base, provisioned, target common.FileBuffer, haveForce bool) (skipReport bool, errs []error) {
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	//if the package management left behind additional cleanup targets
	//(most likely a backup of our custom configuration), we can delete
	//these too
	cleanupTargets := platform.Implementation().AdditionalCleanupTargets(target.Path)
	for _, otherFile := range cleanupTargets {
		fmt.Printf(">> also deleting %s\n", otherFile)
		appendError(os.Remove(otherFile))
	}
	appendError(os.Remove(base.Path))
	appendError(os.Remove(provisioned.Path))
	return false, errs
}

func (tf *TargetFile) applyRestore(base, provisioned, target common.FileBuffer, haveForce bool) (skipReport bool, errs []error) {
	appendError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	//target is still there - restore the target base, *but* before that,
	//check if there is an updated target base
	updatedTBPath, reportedTBPath, err := platform.Implementation().FindUpdatedTargetBase(target.Path)
	appendError(err)
	if updatedTBPath != "" {
		fmt.Printf(">> found updated target base: %s -> %s\n", reportedTBPath, target.Path)
		//use this target base instead of the one in the TargetBaseDirectory
		appendError(os.Remove(base.Path))
		base.Path = updatedTBPath
	}

	//now really restore the target base
	appendError(common.MoveFile(base.Path, target.Path))
	appendError(os.Remove(provisioned.Path))
	return false, errs
}
