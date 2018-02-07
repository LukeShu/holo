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

// Command holo is the user interface to Holo.
package entrypoint

import (
	"fmt"
	"io"
	"os"

	"github.com/holocm/holo/cmd/holo/internal/impl"
	"github.com/holocm/holo/cmd/holo/internal/output"
)

//this is populated at compile-time, see Makefile
var version = "unknown"

// Main is the main entry point, but returns the exit code rather than
// calling os.Exit().  This distinction is useful for monobinary and
// testing purposes.
func Main() (exitCode int) {
	//a command word must be given as first argument
	if len(os.Args) < 2 {
		help(os.Stderr)
		return 2
	}

	//check that it is a known command word
	var command func([]*impl.EntityHandle, map[int]bool) int
	knownOpts := make(map[string]int)
	switch os.Args[1] {
	case "apply":
		knownOpts = map[string]int{"-f": impl.OptionApplyForce, "--force": impl.OptionApplyForce}
		command = impl.Apply
	case "diff":
		command = impl.Diff
	case "scan":
		knownOpts = map[string]int{
			"-s": impl.OptionScanShort, "--short": impl.OptionScanShort,
			"-p": impl.OptionScanPorcelain, "--porcelain": impl.OptionScanPorcelain,
		}
		command = impl.Scan
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

	return impl.WithCacheDirectory(func() (exitCode int) {
		//load configuration
		config := impl.ReadConfiguration()
		if config == nil {
			//some fatal error occurred - it was already reported, so just exit
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
		entities, err := impl.GetAllEntities(config.Plugins)
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

	}) //end of WithCacheDirectory
}

func help(w io.Writer) {
	program := os.Args[0]
	fmt.Fprintf(w, "Usage: %s apply [-f|--force] [selector ...]\n", program)
	fmt.Fprintf(w, "   or: %s diff [selector ...]\n", program)
	fmt.Fprintf(w, "   or: %s scan [-s|--short|-p|--porcelain] [selector ...]\n", program)
	fmt.Fprintf(w, "   or: %s version\n", program)
	fmt.Fprintf(w, "   or: %s help\n", program)
	fmt.Fprintf(w, "\nSee `man 8 holo` for details.\n")
}
