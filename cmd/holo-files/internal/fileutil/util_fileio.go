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

package fileutil

import (
	"os"
)

// IsManageableFile returns whether the file can be managed by Holo
// (i.e. is a regular file or a symlink).
func IsManageableFile(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return IsManageableFileInfo(info)
}

// IsManageableFileInfo returns whether the given FileInfo refers to a
// manageable file (i.e. a regular file or a symlink).
func IsManageableFileInfo(info os.FileInfo) bool {
	return info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0
}

// CopyFile copies a regular file or symlink, including the file
// metadata.
func CopyFile(fromPath, toPath string) error {
	buf, err := NewFileBuffer(fromPath)
	if err != nil {
		return err
	}
	if buf.Mode&os.ModeSymlink != 0 && IsManageableFile(toPath) {
		err = os.Remove(toPath)
		if err != nil {
			return err
		}
	}
	return buf.Write(toPath)
}

// MoveFile is like CopyFile, but it removes the fromPath after
// successful copying.
func MoveFile(fromPath, toPath string) error {
	err := CopyFile(fromPath, toPath)
	if err != nil {
		return err
	}
	return os.Remove(fromPath)
}
