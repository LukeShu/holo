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

package files

//TargetDirectory is $HOLO_ROOT_DIR (or "/" if not set).
func (p FilesPlugin) targetDirectory() string {
	return p.Runtime.RootDirPath
}

//ResourceDirectory is $HOLO_RESOURCE_DIR.
func (p FilesPlugin) resourceDirectory() string {
	return p.Runtime.ResourceDirPath
}

//TargetBaseDirectory is $HOLO_STATE_DIR/base.
func (p FilesPlugin) targetBaseDirectory() string {
	return p.Runtime.StateDirPath + "/base"
}

//ProvisionedDirectory is $HOLO_STATE_DIR/provisioned.
func (p FilesPlugin) provisionedDirectory() string {
	return p.Runtime.StateDirPath + "/provisioned"
}
