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
	"bytes"
	"io"
	"strings"

	"holocm.org/lib/holo"
)

func (p *Plugin) HoloInfo() map[string]string {
	var stdout bytes.Buffer
	err := p.Command([]string{"info"}, &stdout, nil, nil).Run()
	if err != nil {
		return nil
	}

	info := make(map[string]string)
	for _, line := range strings.Split(stdout.String(), "\n") {
		// ignore esp. blank lines
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		info[parts[0]] = parts[1]
	}
	return info
}

func (p *Plugin) HoloApply(entityID string, withForce bool, stdout, stderr io.Writer) holo.ApplyResult {
	op := "apply"
	if withForce {
		op = "force-apply"
	}

	// execute apply operation
	fd3text, err := p.RunCommandWithFD3([]string{op, entityID}, stdout, stderr)
	if err != nil {
		return holo.NewApplyError(err)
	}

	var result holo.ApplyResult = holo.ApplyApplied
	if err == nil {
		for _, line := range strings.Split(fd3text, "\n") {
			switch line {
			case "not changed":
				result = holo.ApplyAlreadyApplied
			case "requires --force to overwrite":
				result = holo.ApplyExternallyChanged
			case "requires --force to restore":
				result = holo.ApplyExternallyDeleted
			}
		}
	}
	return result
}

func (p *Plugin) HoloDiff(entityID string, stderr io.Writer) (string, string) {
	fd3text, err := p.RunCommandWithFD3([]string{"diff", entityID}, nil, stderr)
	if err != nil {
		return "", ""
	}

	filenames := strings.Split(fd3text, "\000")

	// Were paths given for diffing?  If not, that's okay, not
	// every plugin knows how to diff.
	if len(filenames) < 2 {
		return "", ""
	}

	return filenames[0], filenames[1]
}
