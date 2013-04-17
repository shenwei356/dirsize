// Copyright 2013 Wei Shen (shenwei356@gmail.com). All rights reserved.
// Use of this source code is governed by a MIT-license
// that can be found in the LICENSE file.

// Summarize size of directories and files in directories.
package main

import (
	"errors"
	"flag"
	"fmt"
	. "github.com/shenwei356/util/bytesize"
	. "github.com/shenwei356/util/sortitem"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	sortByAlphabet bool
	sortBySize     bool
	sortReverse    bool
)

// Parse arguments and show usage.
func init() {
	flag.BoolVar(&sortByAlphabet, "a", false, "sort by Alphabet.")
	flag.BoolVar(&sortBySize, "s", true, "sort by Size.")
	flag.BoolVar(&sortReverse, "r", false, "reverse order while sorting.")
	flag.Usage = func() {
		fmt.Printf("\nUsage: %s [OPTION]... [DIR]...\n\n", os.Args[0])
		fmt.Println("Summarize size of directories and files in directories.")
		fmt.Println("by Wei Shen (shenwei356@gmail.com)\n")
		fmt.Println("OPTION:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
	}
}

func main() {
	for _, arg := range flag.Args() {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			size, info, err := FolderSize(arg)
			if err != nil {
				fmt.Println(err)
			}
			// reverse order while sorting
			if sortReverse {
				if sortByAlphabet { // sort by Alphabet
					sort.Sort(Reverse{ByKey{info}})
				} else { // sort by Size
					sort.Sort(ByValue{info})
				}
			} else {
				if sortByAlphabet {
					sort.Sort(ByKey{info})
				} else {
					sort.Sort(Reverse{ByValue{info}})
				}
			}

			fmt.Printf("\n%s: %v\n", arg, ByteSize(size))
			for _, item := range info {
				fmt.Printf("%v\t%s\n", ByteSize(item.Value), item.Key)
			}

		} else {
			fmt.Printf("\n%s: NOT exists!\n", arg)
		}
	}
}

// Get total size of files in a directory, and store the sizes of first level
// directories and files in a key-value list.
func FolderSize(dirname string) (float64, []Item, error) {
	var size float64 = 0
	var info []Item
	info = make([]Item, 0)

	bytes, err := ioutil.ReadFile(dirname)
	if err == nil {
		size1 := float64(len(bytes))
		info = append(info, Item{dirname, size1})
		return size1, info, nil
	}

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		recover()
		return 0, nil, errors.New("ReadDir Error: " + dirname)
	}
	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		if file.IsDir() {
			size1, _, err := FolderSize(filepath.Join(dirname, file.Name()))
			if err != nil {
				recover()
				return 0, nil, err
			}
			size += size1
			info = append(info, Item{file.Name(), size1})
		} else {
			size1 := float64(file.Size())
			size += size1
			info = append(info, Item{file.Name(), size1})
		}
	}
	return size, info, nil
}
