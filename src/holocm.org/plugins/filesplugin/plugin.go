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

	"holocm.org/lib/holo"
)

// FilesPlugin implements holo.Plugin
type FilesPlugin struct {
	Runtime holo.Runtime
}

func (p FilesPlugin) HoloInfo() map[string]string {
	return map[string]string{
		"MIN_API_VERSION": "3",
		"MAX_API_VERSION": "3",
	}
}

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

func (p FilesPlugin) HoloApply(entityID string, force bool, stdout, stderr io.Writer) holo.ApplyResult {
	e, err := p.getEntity(entityID, stderr)
	if err != nil {
		return holo.NewApplyError(err)
	}
	return e.Apply(force, stdout, stderr)
}

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

func NewFilesPlugin(r holo.Runtime) holo.Plugin {
	return FilesPlugin{r}
}
