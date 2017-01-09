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

// Package filesplugin provides a holo.Plugin to provision files.
package filesplugin

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

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

// HoloApply provisions the given entity.
func (p FilesPlugin) HoloApply(entityID string, force bool, stdout, stderr io.Writer) holo.ApplyResult {
	e, err := p.getEntity(entityID, stderr)
	if err != nil {
		return holo.ApplyError(1)
	}
	return e.(*FilesEntity).Apply(force, stdout, stderr)
}

// HoloDiff returns reference files to compare the (expected state,
// current state) of the given entity.
func (p FilesPlugin) HoloDiff(entityID string, stderr io.Writer) (string, string) {
	selectedEntity, err := p.getEntity(entityID, stderr)
	if err != nil {
		return "", ""
	}
	relPath := selectedEntity.(*FilesEntity).relPath
	new := filepath.Join(p.Runtime.StateDirPath+"/provisioned", relPath)
	cur := filepath.Join(p.Runtime.RootDirPath, relPath)
	return new, cur
}

func (p FilesPlugin) getEntity(entityID string, stderr io.Writer) (holo.Entity, error) {
	entities, err := p.HoloScan(stderr)
	if err != nil {
		return nil, err
	}
	for _, entity := range entities {
		if entity.EntityID() == entityID {
			return entity, nil
		}
	}
	fmt.Fprintf(stderr, "!! unknown entity ID \"%s\"\n", entityID)
	return nil, errors.New("")
}

// NewFilesPlugin creates an instance of FilesPlugin.
func NewFilesPlugin(r holo.Runtime) holo.Plugin {
	return FilesPlugin{r}
}
