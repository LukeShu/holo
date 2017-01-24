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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// FileBuffer represents the contents of a file.  It is used in
// HoloApply() as an intermediary product of application steps.
type FileBuffer struct {
	Path     string
	Mode     os.FileMode
	Uid      int
	Gid      int
	Contents string
}

// NewFileBuffer creates a FileBuffer object by reading the manageable
// file at the given path.  The `follow` argument specifies whether to
// follow symlinks.
func NewFileBuffer(path string, follow bool) (FileBuffer, error) {
	var err error

	var info os.FileInfo
	if follow {
		info, err = os.Stat(path)
	} else {
		info, err = os.Lstat(path)
	}
	if err != nil {
		return FileBuffer{}, err
	}
	stat := info.Sys().(*syscall.Stat_t) // FIXME(majewsky): ugly

	buf := FileBuffer{
		Path: path,
		Mode: info.Mode(),
		Uid:  int(stat.Uid),
		Gid:  int(stat.Gid),
	}

	if buf.Mode&os.ModeSymlink != 0 {
		buf.Contents, err = os.Readlink(path)
		if err != nil {
			return FileBuffer{}, err
		}
	} else if buf.Mode.IsRegular() {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return FileBuffer{}, err
		}
		buf.Contents = string(contents)
	} else {
		return FileBuffer{}, &os.PathError{
			Op:   "holocm.org/plugins/filesplugin.NewFileBuffer",
			Path: path,
			Err:  errors.New("not a manageable file"),
		}
	}

	return buf, nil
}

func (fb FileBuffer) Write(path string) error {
	// Check that we're not attempting to overwrite unmanageable
	// files.
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		// Abort because the target location could not be
		// stat()ed
		return err
	}
	if err == nil {
		if !IsManageableFileInfo(info) {
			return &os.PathError{
				Op:   "holocm.org/plugins/filesplugin.FileBuffer.Write",
				Path: path,
				Err:  errors.New("target exists and is not a manageable file"),
			}
		}
	}

	// Before writing to the target, remove what was there before
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if fb.Mode&os.ModeSymlink == 0 {
		err = ioutil.WriteFile(path, []byte(fb.Contents), fb.Mode)
		if err != nil {
			return err
		}
		// apply mode (necessary because the mode passed to
		// WriteFile is only used if the file does not yet
		// exist).
		err = os.Chmod(path, fb.Mode)
		if err != nil {
			return err
		}
	} else {
		err = os.Symlink(fb.Contents, path)
		if err != nil {
			return err
		}
		// Symlinks don't have mode (on most systems).
	}
	// apply ownership
	return os.Lchown(path, fb.Uid, fb.Gid)
}

// ResolveSymlink takes a FileBuffer that contains a symlink, resolves
// it, and returns a new FileBuffer containing the contents of the
// symlink target.
//
// This operation is used by application strategies that require text
// input.  If the given FileBuffer contains file contents, the same
// FileBuffer is returned unaltered.
//
// It uses the FileBuffer's Path to resolve relative symlinks.
func (fb FileBuffer) ResolveSymlink() (FileBuffer, error) {
	// If the FileBuffer isn't a symlink, we can just return it.
	if fb.Mode&os.ModeSymlink == 0 {
		return fb, nil
	}

	// If the symlink target is relative, resolve it
	target := fb.Contents
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(fb.Path), target)
	}

	return NewFileBuffer(target, true)
}
