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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/holocm/holo/cmd/holo-files/internal/fileutil"
	"github.com/holocm/holo/lib/holo"
)

// applyNonOrphan performs the complete application algorithm for the
// given FilesEntity.  This includes taking a copy of the base if
// necessary, applying all resources, and saving the result in the
// target path with the correct file metadata.
func (entity *FilesEntity) applyNonOrphan(withForce bool, stdout, stderr io.Writer) (holo.ApplyResult, error) {
	// step 1: check if a system update installed a new version of
	// the stock configuration
	//
	// This has to come first because it might shuffle some files
	// around, and if we do anything else first, we might end up
	// stat()ing the wrong file.
	newBasePath, newBase, err := entity.GetNewBase(stdout, stderr)
	if err != nil {
		return nil, err
	}

	// step 2: Load our 3 versions into memory.
	current, err := entity.GetCurrent()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return nil, err
	}

	base, err := entity.GetBase()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return nil, err
	}

	provisioned, err := entity.GetProvisioned()
	if err != nil && !os.IsNotExist(err) {
		if pe, ok := err.(*os.PathError); ok {
			err = errors.New("skipping target: " + pe.Err.Error())
		}
		return nil, err
	}

	////////////////////////////////////////////////////////////////////////

	// step 1: if we don't have a base yet, the file at current
	// *is* the base which we have to copy now
	if !base.Manageable && current.Manageable {
		baseDir := filepath.Dir(base.Path)
		err := os.MkdirAll(baseDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Cannot create directory %s: %s", baseDir, err.Error())
		}

		err = current.Write(base.Path)
		if err != nil {
			return nil, fmt.Errorf("Cannot copy %s to %s: %s", current.Path, base.Path, err.Error())
		}
		tmp := current
		tmp.Path = base.Path
		base = tmp
	}

	if !base.Manageable {
		return nil, errors.New("skipping target: not a manageable file")
	}

	// step 2: make sure there is a current file (unless --force)
	if !current.Manageable {
		if !withForce {
			return holo.ApplyExternallyDeleted, nil
		}
	}

	// step 3: check if a system update installed a new version of
	// the stock configuration
	if newBase.Manageable {
		// an updated stock configuration is available at
		// newBase.Path (but show it to the user as
		// newBasePath)
		fmt.Fprintf(stdout, ">> found updated target base: %s -> %s\n", newBasePath, base.Path)
		err := newBase.Write(base.Path)
		if err != nil {
			return nil, fmt.Errorf("Cannot copy %s to %s: %v", newBase.Path, base.Path, err)
		}
		_ = os.Remove(newBase.Path) // this can fail silently
		newBase.Path = base.Path
		base = newBase
	}

	// step 4: apply the resources *iff* the version at
	// current.Path is the one installed by the package (which can
	// be found at base.Path); complain if the user made any
	// changes to config files governed by holo (this check is
	// overridden by the --force option)

	// render desired state of entity
	desired, err := entity.GetDesired(base, stdout, stderr)
	if err != nil {
		return nil, err
	}

	// compare it against the current expected state (a reference
	// file for this must exist at this point); normally this will
	// be the last-provisioned version, but if we've never
	// provisioned it before, then it is the base version
	expected := provisioned
	if !provisioned.Manageable {
		expected = base
	}
	if !(current.EqualTo(expected) || current.EqualTo(desired)) {
		if !withForce {
			return holo.ApplyExternallyChanged, nil
		}
	}

	// save a copy of the provisioned config file to check for
	// manual modifications in the next Apply() run
	if !desired.EqualTo(provisioned) {
		provisionedDir := filepath.Dir(provisioned.Path)
		err = os.MkdirAll(provisionedDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Cannot write %s: %s", provisioned.Path, err.Error())
		}
		err = desired.Write(provisioned.Path)
		if err != nil {
			return nil, err
		}
	}

	if !desired.EqualTo(current) {
		// write the result buffer to the target and copy
		// owners/permissions from base file to target file
		newTargetPath := current.Path + ".holonew"
		err = desired.Write(newTargetPath)
		if err != nil {
			return nil, err
		}
		// move $target.holonew -> $target atomically (to
		// ensure that there is always a valid file at
		// $target)
		err = os.Rename(newTargetPath, current.Path)
		if err != nil {
			return nil, err
		}
		return holo.ApplyApplied, nil
	}
	return holo.ApplyAlreadyApplied, nil
}

//GetBase return the package manager-supplied base version of the
//entity, as recorded the last time it was provisioned.
func (entity *FilesEntity) GetBase() (fileutil.FileBuffer, error) {
	return fileutil.NewFileBuffer(entity.PathIn(entity.plugin.Runtime.StateDirPath + "/base"))
}

//GetProvisioned returns the recorded last-provisioned state of the
//entity.
func (entity *FilesEntity) GetProvisioned() (fileutil.FileBuffer, error) {
	return fileutil.NewFileBuffer(entity.PathIn(entity.plugin.Runtime.StateDirPath + "/provisioned"))
}

//GetCurrent returns the current version of the entity.
func (entity *FilesEntity) GetCurrent() (fileutil.FileBuffer, error) {
	return fileutil.NewFileBuffer(entity.PathIn(entity.plugin.Runtime.RootDirPath))
}

//GetNewBase returns the base version of the entity, if it has been
//updated by the package manager since last applied.
func (entity *FilesEntity) GetNewBase(stdout, stderr io.Writer) (path string, buf fileutil.FileBuffer, err error) {
	realPath, path, err := GetPackageManager(entity.plugin.Runtime.RootDirPath, stdout, stderr).FindUpdatedTargetBase(entity.PathIn(entity.plugin.Runtime.RootDirPath))
	if err != nil {
		return
	}
	if realPath != "" {
		buf, err = fileutil.NewFileBuffer(realPath)
		return
	}
	return
}

//GetDesired applies all the resources for this FilesEntity onto the base.
func (entity *FilesEntity) GetDesired(base fileutil.FileBuffer, stdout, stderr io.Writer) (fileutil.FileBuffer, error) {
	resources := entity.Resources()

	// Optimization: check if we can skip any application steps
	firstStep := 0
	for idx, resource := range resources {
		if resource.DiscardsPreviousBuffer() {
			firstStep = idx
		}
	}
	resources = resources[firstStep:]

	// load the base into a buffer as the start for the
	// application algorithm
	buffer := base
	buffer.Path = entity.PathIn(entity.plugin.Runtime.RootDirPath)

	// apply all the applicable resources in order
	var err error
	for _, resource := range resources {
		buffer, err = resource.ApplyTo(buffer, stdout, stderr)
		if err != nil {
			return fileutil.FileBuffer{}, err
		}
	}

	return buffer, nil
}
