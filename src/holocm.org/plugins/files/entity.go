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

package files

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"holocm.org/lib/holo"
)

// FilesEntity implements holo.Entity.
//
// It represents a configuration file that can be provisioned by Holo.
type FilesEntity struct {
	relTargetPath string //the target path relative to the plugin.targetDirectory()
	orphaned      bool   //default: false
	repoEntries   RepoFiles
	plugin        FilesPlugin
}

// NewFilesEntityFromPathIn creates a FilesEntity instance for which a
// path relative to a known location is known.
//
//    target := p.NewFilesEntityFromPathIn(p.targetDirectory(), targetPath)
//    target := p.NewFilesEntityFromPathIn(p.provisionedDirectory(), provisionedPath)
func (p FilesPlugin) NewFilesEntityFromPathIn(directory, path string) *FilesEntity {
	//make path relative
	relTargetPath, _ := filepath.Rel(directory, path)
	return &FilesEntity{relTargetPath: relTargetPath, plugin: p}
}

// PathIn returns the path to this target file relative to the given
// directory.
//
//    targetPath := target.pathIn(target.plugin.targetDirectory())           // e.g. "/etc/foo.conf"
//    targetBasePath := target.pathIn(target.plugin.targetBaseDirectory())   // e.g. "/var/lib/holo/files/base/etc/foo.conf"
//    provisionedPath := target.pathIn(target.plugin.provisionedDirectory()) // e.g. "/var/lib/holo/files/provisioned/etc/foo.conf"
func (target *FilesEntity) PathIn(directory string) string {
	return filepath.Join(directory, target.relTargetPath)
}

// AddRepoEntry registers a new repository entry in this FilesEntity
// instance.
func (target *FilesEntity) AddRepoEntry(entry RepoFile) {
	target.repoEntries = append(target.repoEntries, entry)
}

// RepoEntries returns an ordered list of all repository entries for
// this FilesEntity.
func (target *FilesEntity) RepoEntries() []RepoFile {
	sort.Sort(target.repoEntries)
	return target.repoEntries
}

// EntityID returns the entity ID for this target file.
func (target *FilesEntity) EntityID() string {
	return "file:" + target.PathIn(target.plugin.targetDirectory())
}

func (target *FilesEntity) EntityAction() string {
	if target.orphaned {
		_, _, assessment := target.scanOrphanedTargetBase()
		return fmt.Sprintf("Scrubbing (%s)\n", assessment)
	}
	return ""
}

func (target *FilesEntity) EntitySource() []string {
	if target.orphaned {
		return nil
	}
	var ret []string
	for _, entry := range target.RepoEntries() {
		ret = append(ret, entry.Path())
	}
	return ret
}

func (target *FilesEntity) EntityUserInfo() (r []holo.KV) {
	if target.orphaned {
		_, strategy, _ := target.scanOrphanedTargetBase()
		r = append(r, holo.KV{strategy, target.PathIn(target.plugin.targetBaseDirectory())})
	} else {
		r = append(r, holo.KV{"store at", target.PathIn(target.plugin.targetBaseDirectory())})
		for _, entry := range target.RepoEntries() {
			r = append(r, holo.KV{entry.ApplicationStrategy(), entry.Path()})
		}
	}
	return r
}

func (target *FilesEntity) Apply(withForce bool, stdout, stderr io.Writer) holo.ApplyResult {
	// BUG(lukeshu): FilesEntity.Apply: We hide errors here to
	// match the upstream behavior of holo-files:
	// https://github.com/holocm/holo/issues/19
	if target.orphaned {
		errs := target.handleOrphanedTargetBase()
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(stderr, "!! %s\n", err.Error())
			}
			//return holo.ApplyErr(1)
			return holo.ApplyApplied
		}
		return holo.ApplyApplied
	} else {
		result, err := target.apply(withForce)

		if err != nil {
			fmt.Fprintf(stderr, "!! %s\n", err.Error())
			//return holo.ApplyErr(1)
			return holo.ApplyApplied
		}

		return result
	}
}
