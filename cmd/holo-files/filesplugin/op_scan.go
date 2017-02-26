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
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/filesplugin/fileutil"
	"github.com/holocm/holo/lib/holo"
)

// HoloScan returns a slice of all the FilesEntity entities.  The
// entities are guaranteed to have the concrete type "*FileEntity".
// The entities are sorted by path.
func (p FilesPlugin) HoloScan(stderr io.Writer) ([]holo.Entity, error) {
	// walk over the repo to find repo files (and thus the
	// corresponding target files)
	targets := make(map[string]*FilesEntity)
	resourceDir := p.Runtime.ResourceDirPath
	filepath.Walk(resourceDir, func(repoPath string, repoFileInfo os.FileInfo, err error) error {
		// skip over unaccessible stuff
		if err != nil {
			return err
		}
		// only look at manageable files (regular files or
		// symlinks)
		if !fileutil.IsManageableFileInfo(repoFileInfo) {
			return nil
		}
		// don't consider resourceDir itself to be a repo
		// entry (it might have passed the
		// IsManageableFileInfo check because it might be a
		// symlink)
		if repoPath == resourceDir {
			return nil
		}
		// only look at files within subdirectories (files in
		// the repo directory itself are skipped)
		relPath, _ := filepath.Rel(resourceDir, repoPath)
		if !strings.ContainsRune(relPath, filepath.Separator) {
			return nil
		}

		// create new FilesEntity if necessary and store the
		// repo entry in it
		repoEntry := p.NewRepoFile(repoPath)
		targetPath := repoEntry.TargetPath
		if targets[targetPath] == nil {
			targets[targetPath] = p.NewFilesEntity(p.Runtime.RootDirPath, targetPath)
		}
		targets[targetPath].AddRepoEntry(repoEntry)
		return nil
	})

	// walk over the target base directory to find orphaned target
	// bases
	targetBaseDir := p.Runtime.StateDirPath + "/base"
	filepath.Walk(targetBaseDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		// skip over unaccessible stuff
		if err != nil {
			return err
		}
		// only look at manageable files (regular files or
		// symlinks)
		if !fileutil.IsManageableFileInfo(fileInfo) {
			return nil
		}
		// don't consider targetBaseDir itself to be a target
		// base (it might have passed the IsManageableFileInfo
		// check because it might be a symlink)
		if filePath == targetBaseDir {
			return nil
		}

		// check if we have seen the config file for this
		// target base (if not, it's orphaned)
		//
		// TODO(majewsky): s/(targetBase)Path/\1Dir/g and
		// s/(targetBase)File/Path/g
		target := p.NewFilesEntity(targetBaseDir, filePath)
		targetPath := filepath.Join(p.Runtime.RootDirPath, target.relPath)
		if targets[targetPath] == nil {
			target.orphaned = true
			targets[targetPath] = target
		}
		return nil
	})

	// flatten result into list
	result := make([]holo.Entity, 0, len(targets))
	for _, target := range targets {
		result = append(result, target)
	}

	sort.Sort(filesByPath(result))
	return result, nil
}

type filesByPath []holo.Entity

func (f filesByPath) Len() int           { return len(f) }
func (f filesByPath) Less(i, j int) bool { return f[i].EntityID() < f[j].EntityID() }
func (f filesByPath) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }