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
	"fmt"
	"os"

	"github.com/holocm/holo/cmd/holo/internal/output"
)

func CommandApply(entities []*EntityHandle, withForce bool) int {
	for _, entity := range entities {
		entity.Apply(withForce)

		os.Stderr.Sync()
		output.Stdout.EndParagraph()
		os.Stdout.Sync()
	}

	return 0
}

func CommandScan(entities []*EntityHandle, isPorcelain, isShort bool) int {
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

	return 0
}

func CommandDiff(entities []*EntityHandle) int {
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

	return 0
}
