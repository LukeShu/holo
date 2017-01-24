/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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

package common

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

//FileBuffer represents a file, loaded into memory. It is used in holo.Apply() as
//an intermediary product of application steps.
type FileBuffer struct {
	Path     string
	Mode     os.FileMode
	Uid      int
	Gid      int
	Contents string
}

//NewFileBuffer creates a FileBuffer object by reading the manageable file at
//the given path.
func NewFileBuffer(path string) (FileBuffer, error) {
	return newFileBuffer(path, false)
}

func newFileBuffer(path string, follow bool) (FileBuffer, error) {
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

	stat := info.Sys().(*syscall.Stat_t) // UGLY
	fb := FileBuffer{
		Path: path,
		Mode: info.Mode(),
		Uid:  int(stat.Uid),
		Gid:  int(stat.Gid),
	}

	if fb.Mode&os.ModeSymlink != 0 {
		fb.Contents, err = os.Readlink(path)
		if err != nil {
			return FileBuffer{}, err
		}
	} else if fb.Mode.IsRegular() {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return FileBuffer{}, err
		}
		fb.Contents = string(contents)
	} else {
		return FileBuffer{}, &os.PathError{
			Op:   "holo.NewFileBuffer",
			Path: path,
			Err:  errors.New("not a manageable file"),
		}
	}

	return fb, nil
}

func (fb *FileBuffer) Write(path string) error {
	//(check that we're not attempting to overwrite unmanageable files
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		//abort because the target location could not be statted
		return err
	}
	if err == nil {
		if !(info.Mode().IsRegular() || IsFileInfoASymbolicLink(info)) {
			return &os.PathError{
				Op:   "holo.FileBuffer.Write",
				Path: path,
				Err:  errors.New("target exists and is not a manageable file"),
			}
		}
	}

	//before writing to the target, remove what was there before
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	//a manageable file is either a regular file...
	if fb.Mode&os.ModeSymlink == 0 {
		// regular file
		err = ioutil.WriteFile(path, []byte(fb.Contents), fb.Mode)
	} else {
		// symlink
		err = os.Symlink(fb.Contents, path)
	}
	if err != nil {
		return err
	}
	return os.Lchown(path, fb.Uid, fb.Gid)
}

//ResolveSymlink takes a FileBuffer that contains a symlink, resolves it and
//returns a new FileBuffer containing the contents of the symlink target. This
//operation is used by application strategies that require text input. If
//the given FileBuffer contains file contents, the same buffer is returned
//unaltered.
//
//It uses the FileBuffer's Path to resolve relative symlinks.
func (fb FileBuffer) ResolveSymlink() (FileBuffer, error) {
	//if the buffer has contents already, we can use that
	if fb.Mode&os.ModeSymlink == 0 {
		return fb, nil
	}

	//if the symlink target is relative, resolve it
	target := fb.Contents
	if !filepath.IsAbs(target) {
		baseDir := filepath.Dir(fb.Path)
		target = filepath.Join(baseDir, target)
	}

	return newFileBuffer(target, true)
}
