//usr/bin/env go run "$0" "$@"; exit $?

package main

import (
	"bufio"
	"fmt"
	"io"
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
	var mode string
	blocks := map[string]int{}
	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		handleErr(err)
		buf := bufio.NewReader(file)
		err = nil
		for err != io.EOF {
			var line string
			line, err = buf.ReadString('\n')
			if err == io.EOF {
				continue
			}
			handleErr(err)
			line = strings.TrimSuffix(line, "\n")

			if strings.HasPrefix(line, "mode:") {
				if mode == "" {
					mode = line
				} else if mode != line {
					fmt.Fprintf(os.Stderr, "mixed modes: %q != %q\n", mode, line)
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
	}
	keys := make([]string, 0, len(blocks))
	for key := range blocks {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Println(mode)
	for _, block := range keys {
		fmt.Printf("%s %d\n", block, blocks[block])
	}
}
