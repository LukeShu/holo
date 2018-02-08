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

package entrypoint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/holocm/holo/cmd/holo/internal/impl"
	"github.com/holocm/holo/cmd/holo/internal/output"
)

// Package impl implements most of the "holo" program; as a set of
// modular pieces.
//
// Each piece is usable without the others; there is no global or
// shared state between them... mostly.  The "output" package does
// provide some share global state; but it all takes place in the
// "output" package, not here.
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
// All that remains for the calling program to do is option parsing
// and user-interface concerns.
func _Main(rootDir, version string, getPlugin impl.PluginGetter) (exitCode int) {
	const (
		optionApplyForce = iota
		optionScanShort
		optionScanPorcelain
	)

	var runtimeManager *impl.RuntimeManager

	help := func(w io.Writer) {
		program := os.Args[0]
		fmt.Fprintf(w, "Usage: %s apply [-f|--force] [selector ...]\n", program)
		fmt.Fprintf(w, "   or: %s diff [selector ...]\n", program)
		fmt.Fprintf(w, "   or: %s scan [-s|--short|-p|--porcelain] [selector ...]\n", program)
		fmt.Fprintf(w, "   or: %s version\n", program)
		fmt.Fprintf(w, "   or: %s help\n", program)
		fmt.Fprintf(w, "\nSee `man 8 holo` for details.\n")
	}

	// a command word must be given as first argument
	if len(os.Args) < 2 {
		help(os.Stderr)
		return 2
	}

	//check that it is a known command word
	var command func([]*impl.EntityHandle, map[int]bool) int
	knownOpts := make(map[string]int)
	switch os.Args[1] {
	case "apply":
		knownOpts = map[string]int{"-f": optionApplyForce, "--force": optionApplyForce}
		command = func(e []*impl.EntityHandle, options map[int]bool) int {
			pidFile := impl.AcquirePidFile(filepath.Join(rootDir, "run/holo.pid"))
			if pidFile == nil {
				return 255
			}
			defer pidFile.Release()
			return impl.CommandApply(e, options[optionApplyForce])
		}
	case "diff":
		command = func(e []*impl.EntityHandle, options map[int]bool) int {
			return impl.CommandDiff(e)
		}
	case "scan":
		knownOpts = map[string]int{
			"-s": optionScanShort, "--short": optionScanShort,
			"-p": optionScanPorcelain, "--porcelain": optionScanPorcelain,
		}
		command = func(e []*impl.EntityHandle, options map[int]bool) int {
			return impl.CommandScan(e, options[optionScanPorcelain], options[optionScanShort])
		}
	case "version", "--version":
		fmt.Println(version)
		return 0
	case "help", "--help":
		help(os.Stdout)
		return 0
	default:
		help(os.Stderr)
		return 2
	}

	if true {
		// load configuration
		configReader, err := impl.NewConfigReader(rootDir)
		if err != nil {
			output.Errorf(output.Stderr, "%s", err.Error())
			return 255
		}
		config, err := impl.ReadConfig(configReader)
		if err != nil {
			output.Errorf(output.Stderr, "%s", err.Error())
			return 255
		}

		// load plugins
		runtimeManager, err = impl.NewRuntimeManager(rootDir)
		if err != nil {
			return 255
		}
		defer runtimeManager.Close()
		plugins := runtimeManager.GetPlugins(config.Plugins, getPlugin)
		if plugins == nil {
			// some fatal error occurred - it was already
			// reported, so just exit
			return 255
		}

		// parse command line -- classify each argument as either a
		// selector-string or an option-flag.
		args := os.Args[2:]
		options := make(map[int]bool)
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
		entities, err := impl.GetAllEntities(plugins)
		if err != nil {
			output.Errorf(output.Stderr, "%s", err.Error())
			return 255
		}
		if len(selectors) > 0 {
			entities = impl.FilterEntities(entities, selectors)
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
			return 255
		}

		//execute command
		return command(entities, options)
	}
	panic("not reached")
}
