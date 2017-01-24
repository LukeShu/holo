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

package filesplugin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
)

// Holoscript is a Resource that is a script that edits the current
// version of the entity.
type Holoscript struct{ rawResource }

// ApplicationStrategy implements the Resource interface.
func (resource Holoscript) ApplicationStrategy() string { return "passthru" }

// DiscardsPreviousBuffer implements the Resource interface.
func (resource Holoscript) DiscardsPreviousBuffer() bool { return false }

// ApplyTo implements the Resource interface.
func (resource Holoscript) ApplyTo(entityBuffer fileutil.FileBuffer, stdout, stderr io.Writer) (fileutil.FileBuffer, error) {
	// application of a holoscript requires file contents
	entityBuffer, err := entityBuffer.ResolveSymlink()
	if err != nil {
		return fileutil.FileBuffer{}, err
	}

	// run command, fetch result file into out (not into the
	// entity directly, in order not to corrupt the file there if
	// the script run fails)
	var out bytes.Buffer
	cmd := exec.Command(resource.Path)
	cmd.Stdin = strings.NewReader(entityBuffer.Contents)
	cmd.Stdout = &out
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return fileutil.FileBuffer{}, fmt.Errorf("execution of %s failed: %s", resource.Path, err.Error())
	}

	// result is the stdout of the script
	entityBuffer.Mode &^= os.ModeType
	entityBuffer.Contents = out.String()
	return entityBuffer, nil
}
