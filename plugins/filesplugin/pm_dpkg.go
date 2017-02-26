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
	"github.com/holocm/holo/plugins/filesplugin/fileutil"
)

// pmDPKG provides the PackageManager for dpkg-based distributions
// (Debian and derivatives).
type pmDPKG struct{}

func (p pmDPKG) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	dpkgDistPath := targetPath + ".dpkg-dist" //may be an updated target base
	dpkgOldPath := targetPath + ".dpkg-old"   //may be a backup of the last provisioned target when the updated target base is at targetPath

	//if "${target}.dpkg-old" exists, move it back to $target and move the
	//updated target base to "${target}.dpkg-dist" so that the usual application
	//logic can continue
	if fileutil.IsManageableFile(dpkgOldPath) {
		err := fileutil.MoveFile(targetPath, dpkgDistPath) // TODO(lukeshu): os.Rename()?
		if err != nil {
			return "", "", err
		}
		err = fileutil.MoveFile(dpkgOldPath, targetPath) // TODO(lukeshu): os.Rename()?
		if err != nil {
			return "", "", err
		}
		return dpkgDistPath, targetPath + " (with .dpkg-old)", nil
	}

	if fileutil.IsManageableFile(dpkgDistPath) {
		return dpkgDistPath, dpkgDistPath, nil
	}
	return "", "", nil
}

func (p pmDPKG) AdditionalCleanupTargets(targetPath string) []string {
	//not used by dpkg
	return []string{}
}
