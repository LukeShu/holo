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
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/holocm/holo/lib/holo"
)

// FilesEntity implements holo.Entity.
//
// It represents a configuration file that can be provisioned by Holo.
type FilesEntity struct {
	relPath     string // the target path relative to the plugin.Runtime.RootDirPath
	orphaned    bool   // default: false
	repoEntries RepoFiles
	plugin      FilesPlugin
}

// NewFilesEntity creates a FilesEntity instance for which a path
// relative to a known location is known.
//
//    target := p.NewFilesEntity(p.Runtime.RootDirPath, targetPath)
//
//    target := p.NewFilesEntity("/var/lib/holo/files/provisioned", "/var/lib/holo/files/provisioned/etc/locale.conf")
//    target := p.NewFilesEntity(p.Runtime.StateDirPath + "/provisioned", provisionedPath)
func (p FilesPlugin) NewFilesEntity(basedir, entitypath string) *FilesEntity {
	// make path relative
	relPath, _ := filepath.Rel(basedir, entitypath)
	return &FilesEntity{relPath: relPath, plugin: p}
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
	return "file:" + filepath.Join("/", target.relPath)
}

func (target *FilesEntity) EntityAction() (verb, reason string) {
	if target.orphaned {
		_, _, assessment := target.scanOrphanedTargetBase()
		return "Scrubbing", assessment
	}
	return "", ""
}

func (target *FilesEntity) EntitySource() []string {
	if target.orphaned {
		return nil
	}
	var ret []string
	for _, entry := range target.RepoEntries() {
		ret = append(ret, entry.Path)
	}
	return ret
}

func (target *FilesEntity) EntityUserInfo() (r []holo.KV) {
	if target.orphaned {
		_, strategy, _ := target.scanOrphanedTargetBase()
		r = append(r, holo.KV{strategy, filepath.Join(target.plugin.Runtime.StateDirPath+"/base", target.relPath)})
	} else {
		r = append(r, holo.KV{"store at", filepath.Join(target.plugin.Runtime.StateDirPath+"/base", target.relPath)})
		for _, entry := range target.RepoEntries() {
			if strings.HasSuffix(entry.Path, ".holoscript") {
				r = append(r, holo.KV{"passthru", entry.Path})
			} else {
				r = append(r, holo.KV{"apply", entry.Path})
			}
		}
	}
	return r
}

func (target *FilesEntity) Apply(withForce bool, stdout, stderr io.Writer) holo.ApplyResult {
	// BUG(lukeshu): FilesEntity.Apply: We hide errors here to
	// match the upstream behavior of holo-files:
	// https://github.com/holocm/holo/issues/19
	if target.orphaned {
		errs := target.handleOrphanedTargetBase(stdout, stderr)
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(stderr, "!! %s\n", err.Error())
			}
			//return holo.NewApplyError(errs[0])
			return holo.ApplyApplied
		}
		return holo.ApplyApplied
	} else {
		result, err := target.apply(withForce, stdout, stderr)

		if err != nil {
			fmt.Fprintf(stderr, "!! %s\n", err.Error())
			//return holo.NewApplyError(err)
			return holo.ApplyApplied
		}

		return result
	}
}
