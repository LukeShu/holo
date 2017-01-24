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
	"path/filepath"
	"strings"
)

// RepoFile represents a single file in the configuration repository.
// The string stored in it is the path to the repo file (also
// accessible as Path()).
type RepoFile struct {
	Path string

	TargetPath    string
	Disambiguator string
}

// NewRepoFile creates a RepoFile instance when its path in the file
// system is known.
func (p FilesPlugin) NewRepoFile(path string) RepoFile {
	relPath, _ := filepath.Rel(p.Runtime.ResourceDirPath, strings.TrimSuffix(path, ".holoscript"))
	segments := strings.SplitN(relPath, string(filepath.Separator), 2)

	return RepoFile{
		Path:          path,
		TargetPath:    filepath.Join(p.Runtime.RootDirPath, segments[1]),
		Disambiguator: segments[0],
	}
}

// DiscardsPreviousBuffer indicates whether applying this file will
// discard the previous file buffer (and thus the effect of all
// previous application steps).  This is used as a hint by the
// application algorithm to decide whether application steps can be
// skipped completely.
func (file RepoFile) DiscardsPreviousBuffer() bool {
	return !strings.HasSuffix(file.Path, ".holoscript")
}

// RepoFiles holds a slice of RepoFile instances, and implements some
// methods to satisfy the sort.Interface interface.
type RepoFiles []RepoFile

func (f RepoFiles) Len() int           { return len(f) }
func (f RepoFiles) Less(i, j int) bool { return f[i].Disambiguator < f[j].Disambiguator }
func (f RepoFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
