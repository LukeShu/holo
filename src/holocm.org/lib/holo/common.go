/*******************************************************************************
*
* Copyright 2015 Stefan Majewsky <majewsky@gmx.net>
* Copyright 2017 Luke Shumaker <lukeshu@parabola.nu>
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

// Package holo provides an interface for plugins.
package holo

import (
	"io"
)

type Runtime struct {
	APIVersion int

	RootDirPath string

	ResourceDirPath string
	StateDirPath    string
	CacheDirPath    string
}

type KV struct {
	Key, Val string
}

type Entity interface {
	EntityID() string
	EntitySource() []string
	EntityAction() (verb, reason string)
	EntityUserInfo() []KV
}

type Plugin interface {
	// Return metadata about the plugin itself.
	//
	// "MIN_API_VERSION" and "MAX_API_VERSION"
	HoloInfo() map[string]string

	// Scan Runtime.ResourceDirPath and return a list of entities
	// that this plugin can provision.
	//
	// Errors are reported immediately and will result in a nil
	// slice being returned.
	HoloScan(stderr io.Writer) ([]Entity, error)

	// Provision entityID
	HoloApply(entityID string, force bool, stdout, stderr io.Writer) ApplyResult

	// Return two file paths.  The file pointed to by the first is
	// a representation of entity in the desired provisioned
	// state; the second is a representation of the entity in its
	// current state.
	//
	// If either state is "not existing", an empty string may be
	// returned for one state.  If the entity does not have a
	// meaningful textual representation, then two empty strings
	// should be returned.
	HoloDiff(entityID string, stderr io.Writer) (string, string)
}
