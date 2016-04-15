/*******************************************************************************
*
* Copyright 2015-2016 Stefan Majewsky <majewsky@gmx.net>
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

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	etcPasswdPath string
	etcGroupPath  string
)

func init() {
	rootDir := os.Getenv("HOLO_ROOT_DIR")
	if rootDir == "" {
		rootDir = "/"
	}
	etcPasswdPath = filepath.Join(rootDir, "etc/passwd")
	etcGroupPath = filepath.Join(rootDir, "etc/group")
}

//Getent reads entries from a UNIX user/group database (e.g. /etc/passwd
//or /etc/group) and returns the first entry matching the given predicate.
//For example, to locate the user with name "foo":
//
//    fields, err := Getent("/etc/passwd", func(fields []string) bool {
//        return fields[0] == "foo"
//    })
func Getent(databaseFile string, predicate func([]string) bool) ([]string, error) {
	//read database file
	contents, err := ioutil.ReadFile(databaseFile)
	if err != nil {
		return nil, err
	}

	//each entry is one line
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	for _, line := range lines {
		//fields inside the entries are separated by colons
		fields := strings.Split(strings.TrimSpace(line), ":")
		if predicate(fields) {
			return fields, nil
		}
	}

	//no entry matches
	return nil, nil
}

//GetProvisionedState implements the EntityDefinition interface.
func (g *GroupDefinition) GetProvisionedState() (EntityDefinition, error) {
	//fetch entry from /etc/group
	fields, err := Getent(etcGroupPath, func(fields []string) bool { return fields[0] == g.Name })
	if err != nil {
		return nil, err
	}
	//is there such a group?
	if fields == nil {
		return nil, nil
	}
	//is the group entry intact?
	if len(fields) < 4 {
		return nil, errors.New("invalid entry in /etc/group (not enough fields)")
	}

	//read fields in entry
	gid, err := strconv.Atoi(fields[2])
	return &GroupDefinition{
		Name: fields[0],
		GID:  gid,
	}, err
}

//GetProvisionedState implements the EntityDefinition interface.
func (u *UserDefinition) GetProvisionedState() (EntityDefinition, error) {
	//fetch entry from /etc/passwd
	fields, err := Getent(etcPasswdPath, func(fields []string) bool { return fields[0] == u.Name })
	if err != nil {
		return nil, err
	}
	//is there such a user?
	if fields == nil {
		return nil, nil
	}
	//is the passwd entry intact?
	if len(fields) < 4 {
		return nil, errors.New("invalid entry in /etc/passwd (not enough fields)")
	}

	//read fields in passwd entry
	actualUID, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}

	//fetch entry for login group from /etc/group (to resolve actualGID into a
	//group name)
	actualGIDString := fields[3]
	groupFields, err := Getent(etcGroupPath, func(fields []string) bool {
		if len(fields) <= 2 {
			return false
		}
		return fields[2] == actualGIDString
	})
	if err != nil {
		return nil, err
	}
	if groupFields == nil {
		return nil, errors.New("invalid entry in /etc/passwd (login group does not exist)")
	}
	groupName := groupFields[0]

	//check /etc/group for the supplementary group memberships of this user
	var groupNames []string
	_, err = Getent(etcGroupPath, func(fields []string) bool {
		if len(fields) <= 3 {
			return false
		}
		//collect groups that contain this user
		users := strings.Split(fields[3], ",")
		for _, user := range users {
			if user == u.Name {
				groupNames = append(groupNames, fields[0])
			}
		}
		//keep going
		return false
	})
	if err != nil {
		return nil, err
	}

	//make sure that the groups list is always sorted (esp. for reproducible test output)
	sort.Strings(groupNames)

	return &UserDefinition{
		Name:    fields[0],
		Comment: fields[4],
		UID:     actualUID,
		Home:    fields[5],
		Group:   groupName,
		Groups:  groupNames,
		Shell:   fields[6],
	}, nil
}
