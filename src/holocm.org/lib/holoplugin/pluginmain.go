package holoplugin

import (
	"fmt"
	"os"
	"strconv"

	"holocm.org/lib/holo"
)

func Main(getplugin func(holo.Runtime) holo.Plugin) {
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
			os.Exit(1)
		}
		if runtime.APIVersion < 1 {
			fmt.Fprintf(os.Stderr, "HOLO_API_VERSION must be positive: %d", runtime.APIVersion)
			os.Exit(1)
		}
	}

	plugin := getplugin(runtime)

	switch os.Args[1] {
	case "info":
		for key, val := range plugin.HoloInfo() {
			fmt.Printf("%s=%s\n", key, val)
		}
	case "scan":
		entities, err := plugin.HoloScan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		for _, entity := range entities {
			fmt.Printf("ENTITY: %s\n", entity.EntityID())
			for _, source := range entity.EntitySource() {
				fmt.Printf("SOURCE: %s\n", source)
			}
			if entity.EntityAction() != "" {
				fmt.Printf("ACTION: %s\n", entity.EntityAction())
			}
			for _, kv := range entity.EntityUserInfo() {
				fmt.Printf("%s: %s\n", kv.Key, kv.Val)
			}
		}
	case "apply":
		plugin.HoloApply(os.Args[2], false).Exit()
	case "force-apply":
		plugin.HoloApply(os.Args[2], true).Exit()
	case "diff":
		new, cur := plugin.HoloDiff(os.Args[2])
		if new == "" && cur == "" {
			os.Exit(0)
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
	os.Exit(0)
}
