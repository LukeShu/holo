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
	"errors"
	"fmt"
	"io"
	"os"

	"holocm.org/lib/holo"
)

type FilesPlugin struct {
	Runtime holo.Runtime
}

func (p FilesPlugin) HoloInfo() map[string]string {
	return map[string]string{
		"MIN_API_VERSION": "3",
		"MAX_API_VERSION": "3",
	}
}

func (p FilesPlugin) HoloScan() ([]holo.Entity, error) {
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
	return p.getEntity(entityID).Apply(force, stdout, stderr)
}

func (p FilesPlugin) HoloDiff(entityID string) (string, string) {
	selectedEntity := p.getEntity(entityID)
	new := selectedEntity.PathIn(p.provisionedDirectory())
	cur := selectedEntity.PathIn(p.targetDirectory())
	return new, cur
}

func (p FilesPlugin) getEntity(entityID string) *TargetFile {
	entities := p.ScanRepo()
	if entities == nil {
		// some fatal error occurred - it was already
		// reported, so just exit
		os.Exit(1)
	}
	var selectedEntity *TargetFile
	for _, entity := range entities {
		if entity.EntityID() == entityID {
			selectedEntity = entity
			break
		}
	}
	if selectedEntity == nil {
		fmt.Fprintf(os.Stderr, "!! unknown entity ID \"%s\"\n", entityID)
		os.Exit(1)
	}
	return selectedEntity
}

func NewFilesPlugin(r holo.Runtime) holo.Plugin {
	return FilesPlugin{r}
}
