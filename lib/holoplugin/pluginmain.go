/*******************************************************************************
*
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

// Package holoplugin provides a main() method for turning a
// holo.Plugin into an executable.
package holoplugin

import (
	"fmt"
	"os"
	"strconv"

	"github.com/holocm/holo/lib/holo"
)

func Main(getplugin func(holo.Runtime) holo.Plugin) int {
	runtime := holo.Runtime{
		RootDirPath: os.Getenv("HOLO_ROOT_DIR"),

		ResourceDirPath: os.Getenv("HOLO_RESOURCE_DIR"),
		StateDirPath:    os.Getenv("HOLO_STATE_DIR"),
		CacheDirPath:    os.Getenv("HOLO_CACHE_DIR"),
	}
	if runtime.RootDirPath == "" {
		runtime.RootDirPath = "/"
	}
	if verstr, ok := os.LookupEnv("HOLO_API_VERSION"); ok {
		var err error
		runtime.APIVersion, err = strconv.Atoi(verstr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse HOLO_API_VERSION: %v", err)
			return 1
		}
		if runtime.APIVersion < 1 {
			fmt.Fprintf(os.Stderr, "HOLO_API_VERSION must be positive: %d", runtime.APIVersion)
			return 1
		}
	}

	plugin := getplugin(runtime)

	switch os.Args[1] {
	case "info":
		for key, val := range plugin.HoloInfo() {
			fmt.Printf("%s=%s\n", key, val)
		}
	case "scan":
		entities, err := plugin.HoloScan(os.Stderr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			return 1
		}
		for _, entity := range entities {
			fmt.Printf("ENTITY: %s\n", entity.EntityID())
			for _, source := range entity.EntitySource() {
				fmt.Printf("SOURCE: %s\n", source)
			}
			if verb, reason := entity.EntityAction(); verb != "" {
				if reason == "" {
					fmt.Printf("ACTION: %s\n", verb)
				} else {
					fmt.Printf("ACTION: %s (%s)\n", verb, reason)
				}
			}
			for _, kv := range entity.EntityUserInfo() {
				fmt.Printf("%s: %s\n", kv.Key, kv.Val)
			}
		}
	case "apply":
		result := plugin.HoloApply(os.Args[2], false, os.Stdout, os.Stderr)
		if msg, ok := result.(holo.ApplyMessage); ok {
			msg.Send()
		}
		return result.ExitCode()
	case "force-apply":
		result := plugin.HoloApply(os.Args[2], true, os.Stdout, os.Stderr)
		if msg, ok := result.(holo.ApplyMessage); ok {
			msg.Send()
		}
		return result.ExitCode()
	case "diff":
		new, cur := plugin.HoloDiff(os.Args[2], os.Stderr)
		if new == "" && cur == "" {
			return 0
		}
		if new == "" {
			new = "/dev/null"
		}
		if cur == "" {
			cur = "/dev/null"
		}
		file := os.NewFile(3, "/dev/fd/3")
		_, err := fmt.Fprintf(file, "%s\x00%s\x00", new, cur)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! %s\n", err.Error())
		}
	}
	return 0
}
