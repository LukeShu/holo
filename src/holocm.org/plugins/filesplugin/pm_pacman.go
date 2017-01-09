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
	"holocm.org/plugins/filesplugin/fileutil"
)

// pmPacman provides the PackageManager for pacman/libalpm-based
// distributions (Arch Linux and derivatives).
type pmPacman struct{}

func (p pmPacman) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	pacnewPath := targetPath + ".pacnew"
	if fileutil.IsManageableFile(pacnewPath) {
		return pacnewPath, pacnewPath, nil
	}
	return "", "", nil
}

func (p pmPacman) AdditionalCleanupTargets(targetPath string) []string {
	pacsavePath := targetPath + ".pacsave"
	if fileutil.IsManageableFile(pacsavePath) {
		return []string{pacsavePath}
	}
	return nil
}
