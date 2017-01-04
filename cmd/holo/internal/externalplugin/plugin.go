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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/holocm/holo/cmd/holo/internal/output"
	"github.com/holocm/holo/lib/holo"
)

//PluginAPIVersion is the version of holo-plugin-interface(7) implemented by this.
const PluginAPIVersion = 3

//ErrPluginExecutableMissing indicates that a plugin's executable file is missing.
var ErrPluginExecutableMissing = errors.New("ErrPluginExecutableMissing")

//Plugin describes a plugin executable adhering to the holo-plugin-interface(7).
type Plugin struct {
	id             string
	executablePath string
	metadata       map[string]string //from "info" call
	Runtime        holo.Runtime
}

var _ holo.Plugin = &Plugin{}

//NewPluginWithExecutablePath creates a new Plugin whose executable resides in
//a non-standard location. (This is used exclusively for testing plugins before
//they are installed.)
func NewPluginWithExecutablePath(id string, executablePath string, runtime holo.Runtime) (*Plugin, error) {
	p := &Plugin{id, executablePath, nil, runtime}

	//check if the plugin executable exists
	_, err := os.Stat(executablePath)
	if err != nil {
		if os.IsNotExist(err) {
			output.Errorf(output.Stderr, "%s: file not found", executablePath)
			return nil, ErrPluginExecutableMissing
		}
		return nil, err
	}

	//load metadata with the "info" command
	_ = p.HoloInfo()
	if p.metadata == nil {
		return nil, errors.New("command \"info\" failed")
	}

	//validate metadata
	minVersion, err := strconv.Atoi(p.metadata["MIN_API_VERSION"])
	if err != nil {
		return nil, err
	}
	maxVersion, err := strconv.Atoi(p.metadata["MAX_API_VERSION"])
	if err != nil {
		return nil, err
	}
	if minVersion > PluginAPIVersion || maxVersion < PluginAPIVersion {
		return nil, fmt.Errorf(
			"plugin holo-%s is incompatible with this Holo (plugin min: %d, plugin max: %d, Holo: %d)",
			p.id, minVersion, maxVersion, PluginAPIVersion,
		)
	}

	return p, nil
}

func (p *Plugin) HoloInfo() map[string]string {
	if p.metadata == nil {
		p.metadata = make(map[string]string)
		var buf bytes.Buffer
		err := p.Command([]string{"info"}, &buf, output.Stderr, nil).Run()
		if err != nil {
			return nil
		}
		lines := strings.Split(string(buf.Bytes()), "\n")
		for _, line := range lines {
			//ignore esp. blank lines
			if !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			p.metadata[parts[0]] = parts[1]
		}
	}
	return p.metadata
}

//ID returns the plugin ID.
func (p *Plugin) ID() string {
	return p.id
}

//Command returns an os.exec.Command structure that is set up to run the plugin
//with the given arguments, producing output on the given output and error
//channels. For commands that use file descriptor 3 as an extra output channel,
//the `msg` file can be given (nil is acceptable too).
//
//Note that if a write end of an os.Pipe() is passed for `msg`, it must be
//Close()d after the child is Start()ed. Otherwise, reads from the read end
//will block forever.
func (p *Plugin) Command(arguments []string, stdout io.Writer, stderr io.Writer, msg *os.File) *exec.Cmd {
	cmd := exec.Command(p.executablePath, arguments...)
	cmd.Stdin = nil
	cmd.Stdout = stdout
	cmd.Stderr = &output.LineColorizingWriter{Writer: stderr, Rules: []output.LineColorizingRule{
		{[]byte("!! "), []byte("\x1B[1;31m")},
		{[]byte(">> "), []byte("\x1B[1;33m")},
	}}
	if msg != nil {
		cmd.ExtraFiles = []*os.File{msg}
	}

	//setup environment
	env := os.Environ()
	env = append(env, "HOLO_API_VERSION="+strconv.Itoa(PluginAPIVersion))
	env = append(env, "HOLO_CACHE_DIR="+filepath.Clean(p.Runtime.CacheDirPath))
	env = append(env, "HOLO_RESOURCE_DIR="+filepath.Clean(p.Runtime.ResourceDirPath))
	env = append(env, "HOLO_STATE_DIR="+filepath.Clean(p.Runtime.StateDirPath))
	if os.Getenv("HOLO_ROOT_DIR") == "" {
		env = append(env, "HOLO_ROOT_DIR="+filepath.Clean(p.Runtime.RootDirPath))
	}
	cmd.Env = env

	return cmd
}

//RunCommandWithFD3 extends the Command function with automatic setup and
//reading of the file-descriptor 3, that is used by some plugin commands to
//report structured messages to Holo.
func (p *Plugin) RunCommandWithFD3(arguments []string, stdout, stderr io.Writer) (string, error) {
	//the command channel (file descriptor 3 on the side of the plugin) can
	//only be set up with an *os.File instance, so use a pipe that the plugin
	//writes into and that we read from
	cmdReader, cmdWriterForPlugin, err := os.Pipe()
	if err != nil {
		return "", err
	}

	//execute apply operation
	cmd := p.Command(arguments, stdout, stderr, cmdWriterForPlugin)
	err = cmd.Start() //cannot use Run() since we need to read from the pipe before the plugin exits
	if err != nil {
		return "", err
	}

	cmdWriterForPlugin.Close() //or next line will block (see Plugin.Command docs)
	cmdBytes, err := ioutil.ReadAll(cmdReader)
	if err != nil {
		return "", err
	}
	err = cmdReader.Close()
	if err != nil {
		return "", err
	}
	return string(cmdBytes), cmd.Wait()
}

// HoloApply provisions the given entity.
func (p *Plugin) HoloApply(entityID string, withForce bool, stdout, stderr io.Writer) holo.ApplyResult {
	command := "apply"
	if withForce {
		command = "force-apply"
	}

	// execute apply operation
	cmdText, err := p.RunCommandWithFD3([]string{command, entityID}, stdout, stderr)
	if err != nil {
		output.Errorf(stderr, err.Error())
		return holo.ApplyError(1)
	}

	var result holo.ApplyResult = holo.ApplyApplied
	if err == nil {
		cmdLines := strings.Split(cmdText, "\n")
		for _, line := range cmdLines {
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

// HoloDiff returns reference files to compare the (expected state,
// current state) of the given entity.
func (p *Plugin) HoloDiff(entityID string, stderr io.Writer) (string, string) {
	cmdText, err := p.RunCommandWithFD3([]string{"diff", entityID}, output.Stdout, output.Stderr)
	if err != nil {
		return "", ""
	}

	// were paths given for diffing? if not, that's okay, not
	// every plugin knows how to diff
	cmdLines := strings.Split(cmdText, "\000")
	if len(cmdLines) < 2 {
		return "", ""
	}

	return cmdLines[0], cmdLines[1]
}
