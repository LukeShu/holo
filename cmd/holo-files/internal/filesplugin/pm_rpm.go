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
	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
)

// pmRPM provides the PackageManager for RPM-based distributions.
type pmRPM struct{}

func (p pmRPM) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	rpmnewPath := targetPath + ".rpmnew"   //may be an updated target base
	rpmsavePath := targetPath + ".rpmsave" //may be a backup of the last provisioned target when the updated target base is at targetPath

	//if "${target}.rpmsave" exists, move it back to $target and move the
	//updated target base to "${target}.rpmnew" so that the usual application
	//logic can continue
	if fileutil.IsManageableFile(rpmsavePath) {
		err := fileutil.MoveFile(targetPath, rpmnewPath) // TODO(lukeshu): os.Rename()?
		if err != nil {
			return "", "", err
		}
		err = fileutil.MoveFile(rpmsavePath, targetPath) // TODO(lukeshu): os.Rename()?
		if err != nil {
			return "", "", err
		}
		return rpmnewPath, targetPath + " (with .rpmsave)", nil
	}

	if fileutil.IsManageableFile(rpmnewPath) {
		return rpmnewPath, rpmnewPath, nil
	}
	return "", "", nil
}

func (p pmRPM) AdditionalCleanupTargets(targetPath string) []string {
	//not used by RPM
	return []string{}
}
