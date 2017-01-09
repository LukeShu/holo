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

package impl

import (
	"holocm.org/cmd/holo/output"
)

// Ask all plugins to scan for entities
func GetAllEntities(plugins []*PluginHandle) ([]*EntityHandle, error) {
	var allEntities []*EntityHandle
	for _, plugin := range plugins {
		entities, err := plugin.Scan()
		if err != nil {
			return nil, err
		}
		allEntities = append(allEntities, entities...)
		output.Stdout.EndParagraph()
	}

	return allEntities, nil
}

// Go through all entities and selectors, and return the entities that
// are matched by a selector.
//
// The set of selectors is passed in as a map of selector-string =>
// bool.  If a selector is used, the value for that selector in the
// map is set to true.
func FilterEntities(allEntities []*EntityHandle, selectors map[string]bool) []*EntityHandle {
	// Now for an M*N algorithm!  Go through all selectors and
	// entities to find which are matched.
	selectedEntities := make([]*EntityHandle, 0, len(allEntities))
	for _, entity := range allEntities {
		isEntitySelected := false
		for selector, _ := range selectors {
			if entity.MatchesSelector(selector) {
				isEntitySelected = true
				selectors[selector] = true
				// Note: don't break from the
				// selectors loop; we want to look at
				// every selector because this loop
				// also verifies that selectors are
				// valid
			}
		}
		if isEntitySelected {
			selectedEntities = append(selectedEntities, entity)
		}
	}
	return selectedEntities
}
