//usr/bin/env go run "$0" "$@"; exit $?

// Command gocovcat combines multiple go cover runs, and prints the
// result on stdout.
package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func handleErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	modeBool := false
	blocks := map[string]int{}
	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		handleErr(err)
		buf := bufio.NewScanner(file)
		for buf.Scan() {
			line := buf.Text()

			if strings.HasPrefix(line, "mode: ") {
				m := strings.TrimPrefix(line, "mode: ")
				switch m {
				case "set":
					modeBool = true
				case "count", "atomic":
					// do nothing
				default:
					fmt.Fprintf(os.Stderr, "Unrecognized mode: %s\n", m)
					os.Exit(1)
				}
			} else {
				sp := strings.LastIndexByte(line, ' ')
				block := line[:sp]
				cntStr := line[sp+1:]
				cnt, err := strconv.Atoi(cntStr)
				handleErr(err)
				blocks[block] += cnt
			}
		}
		handleErr(buf.Err())
	}
	keys := make([]string, 0, len(blocks))
	for key := range blocks {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	modeStr := "count"
	if modeBool {
		modeStr = "set"
	}
	fmt.Printf("mode: %s\n", modeStr)
	for _, block := range keys {
		cnt := blocks[block]
		if modeBool && cnt > 1 {
			cnt = 1
		}
		fmt.Printf("%s %d\n", block, cnt)
	}
}
