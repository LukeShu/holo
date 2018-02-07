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

package entrypoint

import (
	"fmt"
	"os"

	"github.com/holocm/holo/cmd/holo/internal/impl"
	"github.com/holocm/holo/cmd/holo/internal/output"
)

const (
	optionApplyForce = iota
	optionScanShort
	optionScanPorcelain
)

func commandApply(entities []*impl.EntityHandle, options map[int]bool) (exitCode int) {
	//ensure that we're the only Holo instance
	if !impl.AcquireLockfile() {
		return 255
	}
	defer impl.ReleaseLockfile()

	withForce := options[optionApplyForce]
	for _, entity := range entities {
		entity.Apply(withForce)

		os.Stderr.Sync()
		output.Stdout.EndParagraph()
		os.Stdout.Sync()
	}

	return 0
}

func commandScan(entities []*impl.EntityHandle, options map[int]bool) (exitCode int) {
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

	return 0
}

func commandDiff(entities []*impl.EntityHandle, options map[int]bool) (exitCode int) {
	for _, entity := range entities {
		buf, err := entity.RenderDiff()
		if err != nil {
			output.Errorf(output.Stderr, "cannot diff %s: %s", entity.Entity.EntityID(), err.Error())
		}
		os.Stdout.Write(buf)

		os.Stderr.Sync()
		output.Stdout.EndParagraph()
		os.Stdout.Sync()
	}

	return 0
}
