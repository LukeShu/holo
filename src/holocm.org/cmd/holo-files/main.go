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

// Command holo-files provides a separate executable of filesplugin.
package main

import (
	"os"

	"holocm.org/lib/holoplugin"
	"holocm.org/plugins/filesplugin"
)

func Main() int {
	return holoplugin.Main(filesplugin.NewFilesPlugin)
}

func main() {
	os.Exit(Main())
}
