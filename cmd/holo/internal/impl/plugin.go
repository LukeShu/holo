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
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/holocm/holo/cmd/holo/internal/externalplugin"
	"github.com/holocm/holo/cmd/holo/internal/output"
	"github.com/holocm/holo/lib/holo"
)

//PluginAPIVersion is the version of holo-plugin-interface(7) implemented by this.
const PluginAPIVersion = 3

type PluginHandle struct {
	ID      string
	Plugin  holo.Plugin
	Runtime holo.Runtime
	Info    map[string]string
}

func GetPlugin(id string, arg *string) (*PluginHandle, error) {
	if arg == nil {
		_arg := filepath.Join(RootDirectory(), "usr/lib/holo/holo-"+id)
		arg = &_arg
	}
	runtime := holo.Runtime{
		APIVersion:      PluginAPIVersion,
		RootDirPath:     RootDirectory(),
		ResourceDirPath: filepath.Join(RootDirectory(), "usr/share/holo/"+id),
		CacheDirPath:    filepath.Join(CachePath(), id),
		StateDirPath:    filepath.Join(RootDirectory(), "var/lib/holo/"+id),
	}
	plugin, err := externalplugin.NewExternalPlugin(id, *arg, runtime)
	if err != nil {
		return nil, err
	}
	handle := &PluginHandle{
		ID:      id,
		Plugin:  plugin,
		Runtime: runtime,
		Info:    nil,
	}

	// grab/cache metadata
	handle.Info = handle.Plugin.HoloInfo()
	if handle.Info == nil {
		return nil, fmt.Errorf("plugin holo-%s: \"info\" operation failed", handle.ID)
	}
	err = checkVersion(handle)
	if err != nil {
		return nil, err
	}

	return handle, nil
}

func checkVersion(handle *PluginHandle) error {
	minVersion, err := strconv.Atoi(handle.Info["MIN_API_VERSION"])
	if err != nil {
		return err
	}
	maxVersion, err := strconv.Atoi(handle.Info["MAX_API_VERSION"])
	if err != nil {
		return err
	}
	if minVersion > PluginAPIVersion || maxVersion < PluginAPIVersion {
		return fmt.Errorf(
			"plugin holo-%s is incompatible with this Holo (plugin min: %d, plugin max: %d, Holo: %d)",
			handle.ID, minVersion, maxVersion, PluginAPIVersion,
		)
	}
	return nil
}

func HoloApply(handle *PluginHandle, entity holo.Entity, withForce bool) {
	// track whether the report was already printed
	tracker := &output.PrologueTracker{Printer: func() { PrintReport(entity, true) }}
	stdout := &output.PrologueWriter{Tracker: tracker, Writer: output.Stdout}
	stderr := &output.PrologueWriter{Tracker: tracker, Writer: output.Stderr}

	result := handle.Plugin.HoloApply(entity.EntityID(), withForce, stdout, stderr)

	var showReport bool
	var showDiff bool
	switch result {
	case holo.ApplyApplied:
		showReport = true
		showDiff = false
	case holo.ApplyAlreadyApplied:
		showReport = false
		showDiff = false
	case holo.ApplyExternallyChanged:
		output.Errorf(stderr, "Entity has been modified by user (use --force to overwrite)")
		showReport = false
		showDiff = true
	case holo.ApplyExternallyDeleted:
		output.Errorf(stderr, "Entity has been deleted by user (use --force to restore)")
		showReport = true
		showDiff = false
	default: // holo.ApplyError
		showReport = false
		showDiff = false
	}

	if showReport {
		tracker.Exec()
	}
	if showDiff {
		diff, err := RenderDiff(handle.Plugin, entity.EntityID())
		if err != nil {
			output.Errorf(stderr, err.Error())
			return
		}
		// indent diff
		indent := []byte("    ")
		diff = regexp.MustCompile("(?m:^)").ReplaceAll(diff, indent)
		diff = bytes.TrimSuffix(diff, indent)

		tracker.Exec()
		output.Stdout.EndParagraph()
		output.Stdout.Write(diff)
	}
}
