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
		commandHelp()
		return
	}

	//check that it is a known command word
	var command func([]*impl.EntityHandle, map[int]bool)
	knownOpts := make(map[string]int)
	switch os.Args[1] {
	case "apply":
		command = commandApply
		knownOpts = map[string]int{"-f": optionApplyForce, "--force": optionApplyForce}
	case "diff":
		command = commandDiff
	case "scan":
		command = commandScan
		knownOpts = map[string]int{
			"-s": optionScanShort, "--short": optionScanShort,
			"-p": optionScanPorcelain, "--porcelain": optionScanPorcelain,
		}
	case "version", "--version":
		fmt.Println(version)
		return
	default:
		commandHelp()
		return
	}

	//load configuration
	config := ReadConfiguration()
	if config == nil {
		//some fatal error occurred - it was already reported, so just exit
		Exit(255)
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
	command(entities, options)

	Exit(0)
}

func commandHelp() {
	program := os.Args[0]
	fmt.Printf("Usage: %s <operation> [...]\nOperations:\n", program)
	fmt.Printf("    %s apply [-f|--force] [selector ...]\n", program)
	fmt.Printf("    %s diff [selector ...]\n", program)
	fmt.Printf("    %s scan [-s|--short|-p|--porcelain] [selector ...]\n", program)
	fmt.Printf("\nSee `man 8 holo` for details.\n")
}

func commandApply(entities []*impl.EntityHandle, options map[int]bool) {
	if !AcquireLockfile() {
		Exit(255)
	}
	defer ReleaseLockfile()
	withForce := options[optionApplyForce]
	for _, entity := range entities {
		entity.Apply(withForce)

		os.Stderr.Sync()
		output.Stdout.EndParagraph()
		os.Stdout.Sync()
	}
}

func commandScan(entities []*impl.EntityHandle, options map[int]bool) {
	isPorcelain := options[optionScanPorcelain]
	isShort := options[optionScanShort]
	for _, entity := range entities {
		switch {
		case isPorcelain:
			entity.PrintScanReport()
		case isShort:
			fmt.Println(entity.Entity.EntityID())
		default:
			entity.PrintReport(false)
		}
	}
}

func commandDiff(entities []*impl.EntityHandle, options map[int]bool) {
	for _, entity := range entities {
		dat, err := entity.RenderDiff()
		if err != nil {
			output.Errorf(output.Stderr, "cannot diff %s: %s", entity.Entity.EntityID(), err.Error())
		}
		os.Stdout.Write(dat)

		os.Stderr.Sync()
		output.Stdout.EndParagraph()
		os.Stdout.Sync()
	}
}
