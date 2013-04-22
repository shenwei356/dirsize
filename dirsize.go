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
}

func main() {
	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = append(dirs, "./")
	}
	for _, arg := range dirs {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		// Check file existence
		_, err := os.Stat(arg)
		if err != nil {
			fmt.Println(err)
			continue
		}
		size, info, err := FolderSize(arg, true)
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
	}
}

// Get total size of files in a directory, and store the sizes of first level
// directories and files in a key-value list.
func FolderSize(dirname string, firstLevel bool) (float64, []Item, error) {
	var size float64 = 0
	var info []Item
	if firstLevel {
		info = make([]Item, 0)
	}

	// Check the read permission
	f, err := os.Open(dirname)
	if err != nil {
		recover()
		// open-permission-denied file or directory
		return 0, nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadFile(dirname)
	// read file success
	if err == nil {
		size1 := float64(len(bytes))
		if firstLevel {
			info = append(info, Item{dirname, size1})
		}
		return size1, info, nil
	}

	// it's a directory
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		recover()
		return 0, nil, errors.New("ReadDir Error: " + dirname)
	}
	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		fileFullPath := filepath.Join(dirname, file.Name())

		// file or dir judgement could reduce the compute complexity
		// file is not worthing call FolderSize
		if file.IsDir() {
			size1, _, err := FolderSize(fileFullPath, false)
			if err != nil {
				recover()
				// skip this directory
				fmt.Printf("read permission denied  (dir): %s\n", fileFullPath)
				continue
			}
			size += size1
			if firstLevel {
				info = append(info, Item{file.Name(), size1})
			}
		} else {
			// Check the read permission
			// DO NOT use ioutil.ReadFile, which will exhaust the RAM!!!!
			f2, err := os.Open(fileFullPath)
			if err != nil && os.IsPermission(err) {
				recover()
				// open-permission-denied file
				fmt.Printf("read permission denied (file): %s\n", fileFullPath)
				continue
			}

			// to avoid panic "open two many file"
			// defer df2.Close() did not seccess due to "nil pointer err"
			if f2 != nil {
				f2.Close()
			}

			size1 := float64(file.Size())
			size += size1
			if firstLevel {
				info = append(info, Item{file.Name(), size1})
			}
		}
	}
	return size, info, nil
}
