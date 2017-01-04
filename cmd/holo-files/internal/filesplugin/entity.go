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

	"github.com/holocm/holo/lib/holo"
)

// FilesEntity implements holo.Entity.
//
// It represents a configuration file that can be provisioned by Holo.
type FilesEntity struct {
	relPath   string
	resources Resources
	plugin    FilesPlugin
}

// NewEntity creates a Entity instance for which a path is known.
//
//    entity := p.NewEntity("etc/locale.conf")
func (p FilesPlugin) NewEntity(relPath string) *FilesEntity {
	return &FilesEntity{
		relPath: relPath,
		plugin:  p,
	}
}

// PathIn returns the path to this entity relative to the given
// directory.
//
//    var (
//        targetPath      = entity.pathIn(entity.plugin.targetDirectory())      // e.g. "/etc/foo.conf"
//        basePath        = entity.pathIn(entity.plugin.baseDirectory())        // e.g. "/var/lib/holo/files/base/etc/foo.conf"
//        provisionedPath = entity.pathIn(entity.plugin.provisionedDirectory()) // e.g. "/var/lib/holo/files/provisioned/etc/foo.conf"
//    )
func (entity *FilesEntity) PathIn(directory string) string {
	return filepath.Join(directory, entity.relPath)
}

// AddResource registers a new resource in this FilesEntity instance.
func (entity *FilesEntity) AddResource(entry Resource) {
	entity.resources = append(entity.resources, entry)
}

// Resources returns an ordered list of all resources for this
// FilesEntity.
func (entity *FilesEntity) Resources() []Resource {
	sort.Sort(entity.resources)
	return entity.resources
}

// EntityID returns the entity ID for this entity.
func (entity *FilesEntity) EntityID() string {
	return "file:" + entity.PathIn("/")
}

// EntityAction returns a verb describing the action to be taken when
// applying this entity, and optionally a reason justifying that
// action.
func (entity *FilesEntity) EntityAction() (string, string) {
	if len(entity.resources) == 0 {
		_, _, assessment := entity.scanOrphan()
		return "Scrubbing", assessment
	}
	return "", ""
}

// EntitySource returns a list of resourc filenames that make up the
// entity.
func (entity *FilesEntity) EntitySource() []string {
	if len(entity.resources) == 0 {
		return nil
	}
	var ret []string
	for _, resource := range entity.Resources() {
		ret = append(ret, resource.Path())
	}
	return ret
}

// EntityUserInfo returns a list of key/value pairs that will be shown
// to the user during `holo scan`.
func (entity *FilesEntity) EntityUserInfo() (r []holo.KV) {
	if len(entity.resources) == 0 {
		_, strategy, _ := entity.scanOrphan()
		r = append(r, holo.KV{strategy, entity.PathIn(entity.plugin.baseDirectory())})
	} else {
		r = append(r, holo.KV{"store at", entity.PathIn(entity.plugin.baseDirectory())})
		for _, resource := range entity.Resources() {
			r = append(r, holo.KV{resource.ApplicationStrategy(), resource.Path()})
		}
	}
	return r
}

//Apply applies the entity.
func (entity *FilesEntity) Apply(withForce bool, stdout, stderr io.Writer) holo.ApplyResult {
	// BUG(lukeshu): FilesEntity.Apply: We hide errors here to
	// match the upstream behavior of holo-files:
	// https://github.com/holocm/holo/issues/19
	switch len(entity.resources) {
	case 0:
		errs := entity.applyOrphan()
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(stderr, "!! %s\n", err.Error())
			}
			//return holo.ApplyErr(1)
			return holo.ApplyApplied
		}
		return holo.ApplyApplied
	default:
		result, err := entity.applyNonOrphan(withForce)

		if err != nil {
			fmt.Fprintf(stderr, "!! %s\n", err.Error())
			//return holo.ApplyErr(1)
			return holo.ApplyApplied
		}

		return result
	}
}
