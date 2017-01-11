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
package main

import (
	"fmt"
	"io"
	"os"

	"holocm.org/cmd/holo/impl"
	"holocm.org/cmd/holo/output"
)

const (
	optionApplyForce = iota
	optionScanShort
	optionScanPorcelain
)

func main() {
	//a command word must be given as first argument
	if len(os.Args) < 2 {
		help(os.Stderr)
		Exit(2)
	}

	//check that it is a known command word
	var command func([]*impl.EntityHandle)
	knownOpts := make(map[string]int)
	options := make(map[int]bool)
	switch os.Args[1] {
	case "apply":
		knownOpts = map[string]int{"-f": optionApplyForce, "--force": optionApplyForce}
		command = func(e []*impl.EntityHandle) {
			if !AcquireLockfile() {
				Exit(255)
			}
			defer ReleaseLockfile()
			impl.Apply(e, options[optionApplyForce])
		}
	case "diff":
		command = impl.Diff
	case "scan":
		knownOpts = map[string]int{
			"-s": optionScanShort, "--short": optionScanShort,
			"-p": optionScanPorcelain, "--porcelain": optionScanPorcelain,
		}
		command = func(e []*impl.EntityHandle) {
			impl.Scan(e, options[optionScanPorcelain], options[optionScanShort])
		}
	case "version", "--version":
		fmt.Println(version)
		Exit(0)
	case "help", "--help":
		help(os.Stdout)
		Exit(0)
	default:
		help(os.Stderr)
		Exit(2)
	}

	// load configuration
	configReader, err := NewConfigReader(RootDirectory())
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		Exit(255)
	}
	config, err := ReadConfig(configReader)
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		Exit(255)
	}
	plugins := GetPlugins(config.Plugins)
	if plugins == nil {
		// some fatal error occurred - it was already
		// reported, so just exit
		Exit(255)
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
	entities, err := impl.GetAllEntities(plugins)
	if err != nil {
		output.Errorf(output.Stderr, "%s", err.Error())
		Exit(255)
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
		Exit(255)
	}

	//execute command
	command(entities)

	Exit(0)
}

func help(w io.Writer) {
	program := os.Args[0]
	fmt.Fprintf(w, "Usage: %s <operation> [...]\nOperations:\n", program)
	fmt.Fprintf(w, "    %s apply [-f|--force] [selector ...]\n", program)
	fmt.Fprintf(w, "    %s diff [selector ...]\n", program)
	fmt.Fprintf(w, "    %s scan [-s|--short|-p|--porcelain] [selector ...]\n", program)
	fmt.Fprintf(w, "    %s --help\n", program)
	fmt.Fprintf(w, "    %s --version\n", program)
	fmt.Fprintf(w, "\nSee `man 8 holo` for details.\n")
}
