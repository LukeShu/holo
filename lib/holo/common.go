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

// Package holo provides an interface for Holo plugins.
//
// The goal of this package is to provide an abstraction point for
// both the plugins and the plugin-calling-frontend over the
// holo-plugin-interface(7).
package holo

import (
	"io"
)

// Runtime is the context that a plugin runs in.
type Runtime struct {
	APIVersion int

	RootDirPath string

	// The following directories are expected to exist within
	// RootDirPath.  No special adjustment is needed to map them
	// to RootDirPath.
	ResourceDirPath string
	StateDirPath    string
	CacheDirPath    string
}

// Plugin is an interface describing the holo-plugin-interface(7) in
// terms of Go language constructs.
//
// It is expected that an implementor of this takes a Runtime as a
// constructor argument.
type Plugin interface {
	// HoloInfo returns metadata about the plugin itself.
	//
	// It is required that the returned map have keys for
	// "MIN_API_VERSION" and "MAX_API_VERSION"; these two keys
	// describe an interval of versions of the
	// holo-plugin-interface(7) that the plugin is compatible
	// with.  Both values may be identical, of course.
	HoloInfo() map[string]string

	// HoloScan scans Runtime.ResourceDirPath and returns a list
	// of entities that the plugin can provision.
	//
	// Errors should be reported immediately on the Writer passed
	// in as an argument, and should result in a nil slice being
	// returned.
	HoloScan(stderr io.Writer) ([]Entity, error)

	// HoloApply provisions the entity with the given ID.
	//
	// The `force` argument specifies whether the plugin should be
	// forceful when provisioning the entity.
	//
	// Informational output should be printed on the `stdout`
	// Writer, and errors and warnings should be printed on the
	// `stderr` Writer.
	HoloApply(entityID string, force bool, stdout, stderr io.Writer) ApplyResult

	// HoloDiff returns two file paths.  The file pointed to by
	// the first is a representation of entity in the desired
	// fully-provisioned state; the second is a representation of
	// the entity in its current state.
	//
	// If either state is "not existing", an empty string may be
	// returned for one state.  If the entity does not have a
	// meaningful textual representation, then two empty strings
	// should be returned.
	//
	// For entities that are not backed by a file, the plugin is
	// allowed to make up a useful textual representation of the
	// entity, and write appropriate files to the
	// Runtime.CacheDirPath.
	HoloDiff(entityID string, stderr io.Writer) (string, string)
}

// KV is a simple struct for storing a key/value pair.  A list of
// these is useful in place of a map for instances where there may be
// duplicate keys, or when order matters.
type KV struct {
	Key, Val string
}

// Entity is an interface describing an entity that may be provisioned
// by a Holo plugin.
type Entity interface {
	EntityID() string
	EntitySource() []string
	EntityAction() (verb, reason string)
	EntityUserInfo() []KV
}
