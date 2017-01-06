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

package external

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"holocm.org/lib/holo"
)

var scanParseLineRegex = regexp.MustCompile(`^\s*([^:]+): (.+)\s*$`)

func scanParseLine(line string) (key, val string, err error) {
	match := scanParseLineRegex.FindStringSubmatch(line)
	if match == nil {
		return "", "", fmt.Errorf("parse error (line was \"%s\")", line)
	}
	return match[1], match[2], nil
}

var scanParseActionRegex = regexp.MustCompile(`^([^()]+) \((.+)\)$`)

func scanParseAction(action string) (verb, reason string) {
	match := scanParseActionRegex.FindStringSubmatch(action)
	if match == nil {
		return action, ""
	} else {
		return match[1], match[2]
	}
}

// Scan discovers entities available for the given entity. Errors are
// reported immediately and will result in nil being returned. "No
// entities found" will be reported as a non-nil empty slice.
func (p *Plugin) HoloScan(stderr io.Writer) ([]holo.Entity, error) {
	var stdout bytes.Buffer
	err := p.Command([]string{"scan"}, &stdout, stderr, nil).Run()
	if err != nil {
		return nil, fmt.Errorf("scan with plugin %s failed: %s", p.id, err.Error())
	}

	// parse stdout
	result := []holo.Entity{} // non-nil
	var currentEntity *Entity // a pointer to the last element of result
	for idx, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		//skip empty lines
		if line == "" {
			continue
		}

		// keep format strings from getting too long
		errorIntro := fmt.Sprintf("error in scan report of %s, line %d", p.id, idx+1)

		key, value, err := scanParseLine(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", errorIntro, err.Error())
		}

		if currentEntity == nil && key != "ENTITY" {
			return nil, fmt.Errorf("%s: expected entity ID, found attribute \"%s\"", errorIntro, line)
		}

		switch key {
		case "ENTITY":
			// starting new entity
			currentEntity = &Entity{Plugin: p, id: value, actionVerb: "Working on"}
			result = append(result, currentEntity)
		case "SOURCE":
			currentEntity.sourceFiles = append(currentEntity.sourceFiles, value)
		case "ACTION":
			// parse action verb/reason
			currentEntity.actionVerb, currentEntity.actionReason = scanParseAction(value)
		default:
			//store unrecognized keys as info lines
			currentEntity.infoLines = append(currentEntity.infoLines, holo.KV{key, value})
		}
	}

	sort.Sort(entitiesByID(result))
	return result, nil
}

type entitiesByID []holo.Entity

func (e entitiesByID) Len() int           { return len(e) }
func (e entitiesByID) Less(i, j int) bool { return e[i].EntityID() < e[j].EntityID() }
func (e entitiesByID) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
