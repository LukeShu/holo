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

// Package externalplugin provides an implementation of holo.Plugin
// for plugins implemented in external programs via the
// holo-plugin-interface(7).
package externalplugin

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"holocm.org/lib/colorize"
	"holocm.org/lib/holo"
)

// Plugin describes a plugin executable adhering to the
// holo-plugin-interface(7).
type Plugin struct {
	id             string
	executablePath string
	runtime        holo.Runtime
}

var _ holo.Plugin = &Plugin{}

// NewExternalPlugin creates a new Plugin that is implemented in a
// separate executable.
func NewExternalPlugin(id string, executablePath string, runtime holo.Runtime) (holo.Plugin, error) {
	p := &Plugin{
		id:             id,
		executablePath: executablePath,
		runtime:        runtime,
	}

	// check if the plugin executable exists
	_, err := os.Stat(executablePath)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Command returns an os/exec.Command structure that is set up to run
// the plugin with the given arguments, producing output on the given
// output and error channels.  For commands that use file descriptor 3
// as an extra output channel, the `msg` file can be given (nil is
// acceptable too).
func (p *Plugin) command(args []string, fd1 io.Writer, fd2 io.Writer, fd3 *os.File) *exec.Cmd {
	cmd := exec.Command(p.executablePath, args...)
	cmd.Stdin = nil
	cmd.Stdout = fd1
	cmd.Stderr = &colorize.LineColorizingWriter{Writer: fd2, Rules: []colorize.LineColorizingRule{
		{[]byte("!! "), []byte("\x1B[1;31m")},
		{[]byte(">> "), []byte("\x1B[1;33m")},
	}}
	if fd3 != nil {
		cmd.ExtraFiles = []*os.File{fd3}
	}

	//setup environment
	env := os.Environ()
	env = append(env, "HOLO_API_VERSION="+strconv.Itoa(p.runtime.APIVersion))
	env = append(env, "HOLO_CACHE_DIR="+filepath.Clean(p.runtime.CacheDirPath))
	env = append(env, "HOLO_RESOURCE_DIR="+filepath.Clean(p.runtime.ResourceDirPath))
	env = append(env, "HOLO_STATE_DIR="+filepath.Clean(p.runtime.StateDirPath))
	env = append(env, "HOLO_ROOT_DIR="+filepath.Clean(p.runtime.RootDirPath))
	cmd.Env = env

	return cmd
}

// RunCommandWithFD3 extends the Command function with automatic setup
// and reading of the file-descriptor 3, that is used by some plugin
// commands to report structured messages to Holo.
//
// Returns the data written to fd3.
func (p *Plugin) runCommandWithFD3(args []string, stdout, stderr io.Writer) (string, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return "", err
	}

	cmd := p.command(args, stdout, stderr, pipeWriter)

	// cannot use Run() since we need to read from the pipe before the plugin exits
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	pipeWriter.Close()

	fd3bytes, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		return "", err
	}
	err = pipeReader.Close()
	if err != nil {
		return "", err
	}
	return string(fd3bytes), cmd.Wait()
}
