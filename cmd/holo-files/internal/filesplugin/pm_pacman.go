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
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
)

// pmPacman provides the PackageManager for Pacman/libALPM-based
// distributions (Arch Linux and derivatives).
type pmPacman struct{}

func (p pmPacman) FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error) {
	pacnewPath := targetPath + ".pacnew"
	if fileutil.IsManageableFile(pacnewPath) {
		return pacnewPath, pacnewPath, nil
	}
	return "", "", nil
}

func (p pmPacman) AdditionalCleanupTargets(targetPath string) (ret []string) {
	pacsavePath := targetPath + ".pacsave"
	if fileutil.IsManageableFile(pacsavePath) {
		ret = append(ret, pacsavePath)
	}

	// check for .pacsave.+[0-9]
	dir := filepath.Dir(targetPath)
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	base := filepath.Base(targetPath) + ".pacsave."
	for _, fileinfo := range fileinfos {
		if !strings.HasPrefix(fileinfo.Name(), base) {
			continue
		}
		suffix := strings.TrimPrefix(fileinfo.Name(), base)
		if _, err := strconv.ParseUint(suffix, 10, 0); err != nil {
			continue
		}
		if !fileutil.IsManageableFileInfo(fileinfo) {
			continue
		}
		ret = append(ret, filepath.Join(dir, fileinfo.Name()))
	}

	return
}
