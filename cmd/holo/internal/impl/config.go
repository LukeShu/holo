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
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LineReader interface {
	ReadLine() (string, error)
}

// Don't simply use bufio.Reader(io.MultiReader(...)) because if there
// is a holorc.d file without a trailing newline, that line would get
// merged with the first line of the next file.
type configReader struct {
	files []io.Reader
	buf   *bufio.Reader
}

func NewConfigReader(RootDirectory string) (LineReader, error) {
	// etc/holorc.d/*
	dirPath := filepath.Join(RootDirectory, "etc/holorc.d")
	fis, err := ioutil.ReadDir(dirPath)
	if err != nil && !os.IsNotExist(err) {
		// non-existence of the directory is acceptable
		return nil, err
	}
	var paths []string
	for _, fi := range fis {
		if !fi.Mode().IsDir() {
			paths = append(paths, filepath.Join(dirPath, fi.Name()))
		}
	}
	sort.Strings(paths)
	// etc/holorc
	paths = append(paths, filepath.Join(RootDirectory, "etc/holorc"))
	// open all of them
	var files []io.Reader
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return &configReader{files: files}, nil
}

func (r *configReader) ReadLine() (string, error) {
	if r.buf == nil {
		if len(r.files) == 0 {
			return "", io.EOF
		}
		r.buf = bufio.NewReader(r.files[0])
		r.files = r.files[1:]
	}
	line, err := r.buf.ReadString('\n')
	if err == io.EOF {
		r.buf = nil
		err = nil
	}
	return line, err
}

type PluginConfig struct {
	ID  string
	Arg *string // nullable string
}

func (pc PluginConfig) String() string {
	if pc.Arg == nil {
		return pc.ID
	} else {
		return pc.ID + "=" + *pc.Arg
	}
}

// Configuration contains the parsed contents of `/etc/holorc` and
// `/etc/holorc.d/*`.
type Config struct {
	Plugins []PluginConfig
}

// ReadConfig reads the configuration files `/etc/holorc` and
// `/etc/holorc.d/*`.
func ReadConfig(r LineReader) (*Config, error) {
	var result Config
	var err error
	for err != io.EOF {
		// Read a line
		var line string
		line, err = r.ReadLine()
		if err != nil && err != io.EOF {
			return nil, err
		}
		line = strings.TrimSpace(line)

		// Parse the line
		switch {
		case strings.HasPrefix(line, "#") || line == "":
			// skip comments and empty lines
		case strings.HasPrefix(line, "plugin "):
			pluginSpec := strings.TrimSpace(strings.TrimPrefix(line, "plugin"))
			var plugin PluginConfig
			if strings.Contains(pluginSpec, "=") {
				fields := strings.SplitN(pluginSpec, "=", 2)
				plugin.ID = fields[0]
				plugin.Arg = &fields[1]
			} else {
				plugin.ID = pluginSpec
			}
			result.Plugins = append(result.Plugins, plugin)
		default:
			return nil, fmt.Errorf("cannot parse configuration: unknown command: %q", line)
		}
	}

	return &result, nil
}
