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
	"errors"
	"fmt"
	"io"

	"github.com/holocm/holo/lib/holo"
)

// FilesPlugin implements a holo.Plugin for provisioning files.
type FilesPlugin struct {
	Runtime holo.Runtime
}

// HoloInfo returns metadata about this plugin.
func (p FilesPlugin) HoloInfo() map[string]string {
	return map[string]string{
		"MIN_API_VERSION": "3",
		"MAX_API_VERSION": "3",
	}
}

// HoloScan returns a list of entities that this plugin can provision.
func (p FilesPlugin) HoloScan(stderr io.Writer) ([]holo.Entity, error) {
	a := p.ScanRepo()
	if a == nil {
		return nil, errors.New("")
	}
	b := make([]holo.Entity, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b, nil
}

// HoloApply provisions the given entity.
func (p FilesPlugin) HoloApply(entityID string, force bool, stdout, stderr io.Writer) holo.ApplyResult {
	e, err := p.getEntity(entityID, stderr)
	if err != nil {
		return holo.ApplyError(1)
	}
	return e.Apply(force, stdout, stderr)
}

// HoloDiff returns reference files to compare the (expected state,
// current state) of the given entity.
func (p FilesPlugin) HoloDiff(entityID string, stderr io.Writer) (string, string) {
	selectedEntity, err := p.getEntity(entityID, stderr)
	if err != nil {
		return "", ""
	}
	new := selectedEntity.PathIn(p.provisionedDirectory())
	cur := selectedEntity.PathIn(p.targetDirectory())
	return new, cur
}

func (p FilesPlugin) getEntity(entityID string, stderr io.Writer) (*FilesEntity, error) {
	entities := p.ScanRepo()
	if entities == nil {
		// some fatal error occurred - it was already
		// reported, so just exit
		return nil, errors.New("")
	}
	var selectedEntity *FilesEntity
	for _, entity := range entities {
		if entity.EntityID() == entityID {
			selectedEntity = entity
			break
		}
	}
	if selectedEntity == nil {
		fmt.Fprintf(stderr, "!! unknown entity ID \"%s\"\n", entityID)
		return nil, errors.New("")
	}
	return selectedEntity, nil
}

// NewFilesPlugin creates an instance of FilesPlugin.
func NewFilesPlugin(r holo.Runtime) holo.Plugin {
	return FilesPlugin{r}
}
