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

package externalplugin

import (
	"holocm.org/lib/holo"
)

//Entity represents an entity known to some Holo plugin.
type Entity struct {
	Plugin       *Plugin
	id           string
	actionVerb   string
	actionReason string
	sourceFiles  []string
	infoLines    []holo.KV
}

var _ holo.Entity = &Entity{}

func (e *Entity) EntityID() string { return e.id }

func (e *Entity) EntitySource() []string { return e.sourceFiles }

func (e *Entity) EntityUserInfo() []holo.KV { return e.infoLines }

func (e *Entity) EntityAction() (string, string) {
	return e.actionVerb, e.actionReason
}
