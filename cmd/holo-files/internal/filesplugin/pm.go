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

package filesplugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PackageManager provides integration points with a distribution's
// toolchain.
type PackageManager interface {
	// FindUpdatedTargetBase is called as part of the resource
	// application algorithm.  If the system package manager
	// updates a file which has been modified by Holo, it will
	// usually place the new stock configuration next to the
	// targetPath (usually with a special suffix).  If such a file
	// exists, this method must return its name, so that Holo can
	// pick it up and use it as a new base configuration.
	//
	// The reportedPath is usually the same as the actualPath, but
	// some implementations have to move files around, in which
	// case the reportedPath is the original path to the updated
	// target base, and the actualPath is where Holo will find the
	// file.
	FindUpdatedTargetBase(targetPath string) (actualPath, reportedPath string, err error)

	// AdditionalCleanupTargets is called as part of the orphan
	// handling.  When an application package is removed, but one
	// of its configuration files has been modified by Holo, the
	// system package manager will usually retain a copy next to
	// the targetPath (usually with a special suffix).  If such a
	// file exists, this method must return its name, for Holo to
	// clean it up.
	AdditionalCleanupTargets(targetPath string) []string
}

var pm PackageManager

// GetPackageManager returns the most suitable PackageManager
// implementation for the current system.
func GetPackageManager(rootDir string) PackageManager {
	if pm == nil {
		//which distribution are we running on?
		isDist := getOsRelease(rootDir)
		switch {
		case isDist["alpine"]:
			pm = pmAlpine{}
		case isDist["arch"]:
			pm = pmPacman{}
		case isDist["debian"]:
			pm = pmDPKG{}
		case isDist["fedora"], isDist["suse"]:
			pm = pmRPM{}
		case isDist["unittest"]: // intentionally undocumented
			pm = pmNone{}
		default:
			dists := make([]string, 0, len(isDist))
			for dist := range isDist {
				dists = append(dists, dist)
			}
			sort.Strings(dists)
			fmt.Fprintf(os.Stderr, "!! Running on an unrecognized distribution. Distribution IDs: %s\n", strings.Join(dists, ","))
			fmt.Fprintf(os.Stderr, ">> Please report this error at <https://github.com/holocm/holo/issues/new>\n")
			fmt.Fprintf(os.Stderr, ">> and include the contents of your /etc/os-release file.\n")

			pm = pmNone{}
		}
	}
	return pm
}

// getOsRelease returns a set of distribution IDs, drawing on the ID=
// and ID_LIKE= fields of os-release(5).
func getOsRelease(rootDir string) map[string]bool {
	//read /etc/os-release, fall back to /usr/lib/os-release if not available
	bytes, err := ioutil.ReadFile(filepath.Join(rootDir, "etc/os-release"))
	if err != nil {
		if os.IsNotExist(err) {
			bytes, err = ioutil.ReadFile(filepath.Join(rootDir, "usr/lib/os-release"))
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! Cannot read os-release(5): %v\n", err)
		return nil
	}

	//parse os-release syntax (a harshly limited subset of shell script)
	variables := make(map[string]string)
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		//ignore comments
		if line == "" || line[0] == '#' {
			continue
		}
		//line format is key=value
		if !strings.Contains(line, "=") {
			continue
		}
		split := strings.SplitN(line, "=", 2)
		key, value := split[0], split[1]
		//value may be enclosed in quotes
		switch {
		case strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""):
			value = strings.TrimPrefix(strings.TrimSuffix(value, "\""), "\"")
		case strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'"):
			value = strings.TrimPrefix(strings.TrimSuffix(value, "'"), "'")
		}
		//special characters may be escaped
		value = regexp.MustCompile(`\\(.)`).ReplaceAllString(value, "$1")
		//store assignment
		variables[key] = value
	}

	//the distribution IDs we're looking for are in ID= (single value) or ID_LIKE= (space-separated list)
	result := map[string]bool{variables["ID"]: true}
	if idLike, ok := variables["ID_LIKE"]; ok {
		ids := strings.Split(idLike, " ")
		for _, id := range ids {
			result[id] = true
		}
	}
	return result
}
