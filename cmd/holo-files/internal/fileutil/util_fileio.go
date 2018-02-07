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
	"io/ioutil"
	"os"
	"syscall"
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
	info, err := os.Lstat(fromPath)
	if err != nil {
		return err
	}
	switch info.Mode().IsRegular() {
	case true:
		// regular file
		data, err := ioutil.ReadFile(fromPath)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(toPath, data, 0600)
		if err != nil {
			return err
		}
		return ApplyFilePermissions(fromPath, toPath)
	case false:
		// symbolic link

		//read link target
		target, err := os.Readlink(fromPath)
		if err != nil {
			return err
		}
		//remove old file or link if it exists
		if IsManageableFile(toPath) {
			err = os.Remove(toPath)
			if err != nil {
				return err
			}
		}
		//create new link
		err = os.Symlink(target, toPath)
		if err != nil {
			return err
		}

		return nil
	}
	panic("not reached")
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

// ApplyFilePermissions applies permission flags and ownership from
// the first file to the second file.
func ApplyFilePermissions(fromPath, toPath string) error {
	//apply permissions, ownership, modification date from source file to target file
	//NOTE: We cannot just pass the FileMode in WriteFile(), because its
	//FileMode argument is only applied when a new file is created, not when
	//an existing one is truncated.
	info, err := os.Lstat(fromPath)
	if err != nil {
		return err
	}
	targetInfo, err := os.Lstat(toPath)
	if err != nil {
		return err
	}

	if targetInfo.Mode()&os.ModeSymlink == 0 {
		//apply permissions
		err = os.Chmod(toPath, info.Mode())
		if err != nil {
			return err
		}

		//apply ownership
		stat := info.Sys().(*syscall.Stat_t) // UGLY
		err = os.Chown(toPath, int(stat.Uid), int(stat.Gid))
		if err != nil {
			return err
		}
	}

	return nil
}