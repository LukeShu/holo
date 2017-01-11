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

// Package impl implements most of the "holo" program; as a set of
// modular pieces.
//
// Each piece is usable without the others; there is no global or
// shared state between them... mostly.  The "output" package does
// provide some share global state; but it all takes place in the
// "output" package, not here.
package impl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"holocm.org/cmd/holo/output"
)

// Main ties together all of the other parts of this package, and
// provides a user interface to them.
//
// All you have to do is provide it with:
//
//  1. a root directory path to operate in
//  2. a version string to show to the user
//  3. a function for instantiating plugins
//
// What it does is:
//
// 1. Call "NewConfigReader(rootDir)" to get a LineReader that will
// read lines from all of the relevant configuration files.
//
// 2. Call "ReadConfig(LineReader)" to parse holorc configuration.
//
// 3. Call "NewRuntimeManager(rootDir)" to set up the runtime cache
// directory.
//
// 4. Call "runtimeManager.GetPlugins(config.Plugins,
// YOUR_PLUGIN_LOADER)" to load all of the configued plugins.
// Alternatively, you may loop over config.Plugins yourself, calling
// "NewPluginHandle()" for each plugin, with a runtime you got from
// "runtimeManager.NewRuntime(pluginID)".
//
// 5. Generate a list of entities you want to operate on, represented
// by a "[]*EntityHandle".  You probably want to do this by first
// getting a list of all entities from your plugins, then filtering
// that list to the entities you want.
//
// You can get a list of all entities from your plugins with
// "GetAllEntities(pluginHandles)"; alternatively, you can loop over
// your plugin handles yourself, calling ".Scan()" on each.
//
// Then, you can filter that list with "FilterEntities(entities,
// selectors)"; alternatively, you can loop over your entities and
// selectors calling entityHandle.MatchesSelector(selector) for each.
//
// 6. Call one of "CommandApply", "CommandDiff", or "CommandScan" with
// your list of entities.
//
// Beware that you should probably have a system-wide mutex/lockfile,
// ensuring that only one CommandApply is running on a system at once!
// You can do this with a PID file at a fixed location with
// AcquirePidFile(), and calling pidFile.Release() when done.
//
// Alternatively, you can loop over the list of entities yourself,
// calling ".Apply()", ".PrintReport()", ".PrintScanReport()", or
// ".RenderDiff()" on each.
//
// 7. Call "runtimeManager.Close()" to clean up.
//
//     ----
//
// Most of the complexity of this function is dealing with
// option-parsing and user-interface concerns.
func Main(rootDir, version string, getPlugin PluginGetter) {
	const (
		optionApplyForce = iota
		optionScanShort
		optionScanPorcelain
	)

	var runtimeManager *RuntimeManager

	help := func(w io.Writer) {
		program := os.Args[0]
		fmt.Fprintf(w, "Usage: %s <operation> [...]\nOperations:\n", program)
		fmt.Fprintf(w, "    %s apply [-f|--force] [selector ...]\n", program)
		fmt.Fprintf(w, "    %s diff [selector ...]\n", program)
		fmt.Fprintf(w, "    %s scan [-s|--short|-p|--porcelain] [selector ...]\n", program)
		fmt.Fprintf(w, "    %s --help\n", program)
		fmt.Fprintf(w, "    %s --version\n", program)
		fmt.Fprintf(w, "\nSee `man 8 holo` for details.\n")
	}

	exit := func(code int) {
		if runtimeManager != nil {
			runtimeManager.Close()
		}
		os.Exit(code)
	}

	// a command word must be given as first argument
	if len(os.Args) < 2 {
		help(os.Stderr)
		exit(2)
	}

	//check that it is a known command word
	var command func([]*EntityHandle)
	knownOpts := make(map[string]int)
	options := make(map[int]bool)
	switch os.Args[1] {
	case "apply":
		knownOpts = map[string]int{"-f": optionApplyForce, "--force": optionApplyForce}
		command = func(e []*EntityHandle) {
			pidFile := AcquirePidFile(filepath.Join(rootDir, "run/holo.pid"))
			if pidFile == nil {
				exit(255)
			}
			CommandApply(e, options[optionApplyForce])
			pidFile.Release()
		}
	case "diff":
		command = CommandDiff
	case "scan":
		knownOpts = map[string]int{
			"-s": optionScanShort, "--short": optionScanShort,
			"-p": optionScanPorcelain, "--porcelain": optionScanPorcelain,
		}
		command = func(e []*EntityHandle) {
			CommandScan(e, options[optionScanPorcelain], options[optionScanShort])
		}
	case "version", "--version":
		fmt.Println(version)
		exit(0)
	case "help", "--help":
		help(os.Stdout)
		exit(0)
	default:
		help(os.Stderr)
		exit(2)
	}

	// load configuration
	configReader, err := NewConfigReader(rootDir)
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		exit(255)
	}
	config, err := ReadConfig(configReader)
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		exit(255)
	}

	// load plugins
	runtimeManager, err = NewRuntimeManager(rootDir)
	if err != nil {
		exit(255)
	}
	plugins := runtimeManager.GetPlugins(config.Plugins, getPlugin)
	if plugins == nil {
		// some fatal error occurred - it was already
		// reported, so just exit
		exit(255)
	}

	// parse command line -- classify each argument as either a
	// selector-string or an option-flag.
	args := os.Args[2:]
	selectors := make(map[string]bool)
	for _, arg := range args {
		// either it's a known option for this subcommand...
		if value, ok := knownOpts[arg]; ok {
			options[value] = true
		} else { // ...or it must be a selector
			selectors[arg] = false
		}
	}

	// ask all plugins to scan for entities
	entities, err := GetAllEntities(plugins)
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		exit(255)
	}
	if len(selectors) > 0 {
		entities = FilterEntities(entities, selectors)
	}

	// Were there unrecognized selectors?
	hasUnrecognizedArgs := false
	for selector, recognized := range selectors {
		if !recognized {
			fmt.Fprintf(os.Stderr, "Unrecognized argument: %s\n", selector)
			hasUnrecognizedArgs = true
		}
	}
	if hasUnrecognizedArgs {
		exit(255)
	}

	//execute command
	command(entities)

	exit(0)
}
