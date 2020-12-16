// Copyright 2013-2020 Wei Shen (shenwei356@gmail.com). All rights reserved.
// Use of this source code is governed by a MIT-license
// that can be found in the LICENSE file.

// Summarize size of directories and files in directories.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/shenwei356/util/bytesize"
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
		fmt.Fprintf(os.Stderr, `
dirsize (v1.1)
  Summarize size of directories and files in directories.

Usage: dirsize [OPTION...] [DIR...]

`)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
  Site: https://github.com/shenwei356/dirsize
Author: Wei Shen (shenwei356@gmail.com)

`)
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
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		size, info, err := FolderSize(arg, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// reverse order while sorting
		if !sortReverse {
			if sortByAlphabet { // sort by Alphabet
				sort.Sort(ReverseByKey{info})
			} else { // sort by Size
				sort.Sort(ReverseByValue{info})
			}
		} else {
			if sortByAlphabet {
				sort.Sort(ByKey(info))
			} else {
				sort.Sort(ByValue{info})
			}
		}

		fmt.Printf("\n%s: %v\n", blue(arg), bytesize.ByteSize(size))
		for _, item := range info {
			if item.IsDir {
				fmt.Printf("%10v\t%s\n", bytesize.ByteSize(item.Value), blue(item.Key))
			} else {
				fmt.Printf("%10v\t%s\n", bytesize.ByteSize(item.Value), item.Key)
			}
		}
	}
}

var blue = color.New(color.FgBlue).SprintFunc()

// FolderSize gets total size of files in a directory,
// and stores the sizes of first level
// directories and files in a key-value list.
func FolderSize(dirname string, firstLevel bool) (int64, []Item, error) {
	var size int64 = 0
	var info []Item
	if firstLevel {
		info = make([]Item, 0, 128)
	}

	// Check the read permission
	f, err := os.Open(dirname)
	if err != nil {
		// open-permission-denied file or directory
		return 0, nil, err
	}
	defer f.Close()

	// read info
	fi, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	// it'a a file
	if !fi.IsDir() {
		size1 := fi.Size()
		if firstLevel {
			info = append(info, Item{dirname, size1, false})
		}
		return size1, info, nil
	}

	// it's a directory
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return 0, nil, errors.New("read directory error: " + dirname)
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
				// skip this directory
				fmt.Fprintf(os.Stderr, "read permission denied (dir): %s\n", fileFullPath)
				continue
			}
			size += size1
			if firstLevel {
				info = append(info, Item{file.Name(), size1, true})
			}
		} else {
			mode := file.Mode()
			// ignore pipe file
			if strings.HasPrefix(mode.String(), "p") {
				fmt.Fprintf(os.Stderr, "pipe file ignored: %s\n", fileFullPath)
				continue
			}
			// Check the read permission
			// DO NOT use ioutil.ReadFile, which will exhaust the RAM!!!!
			f2, err := os.Open(fileFullPath)

			if err != nil && os.IsPermission(err) {
				recover()
				// open-permission-denied file
				fmt.Fprintf(os.Stderr, "read permission denied (file): %s\n", fileFullPath)
				continue
			}

			// to avoid panic "open two many file"
			// defer df2.Close() did not seccess due to "nil pointer err"
			if f2 != nil {
				f2.Close()
			}

			size1 := file.Size()
			size += size1
			if firstLevel {
				info = append(info, Item{file.Name(), size1, false})
			}
		}
	}
	return size, info, nil
}

// Item records a file and its size
type Item struct {
	Key   string
	Value int64
	IsDir bool
}

// ByKey sorts by key
type ByKey []Item

func (l ByKey) Len() int           { return len(l) }
func (l ByKey) Less(i, j int) bool { return strings.Compare(l[i].Key, l[j].Key) < 0 }
func (l ByKey) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// ByValue sorts by value
type ByValue struct {
	ByKey
}

// Less checks the order of two element
func (l ByValue) Less(i, j int) bool { return l.ByKey[i].Value < l.ByKey[j].Value }

// ReverseByKey reverses the order
type ReverseByKey struct {
	ByKey
}

// Less checks the order of two element
func (l ReverseByKey) Less(i, j int) bool { return strings.Compare(l.ByKey[i].Key, l.ByKey[j].Key) > 0 }

// ReverseByValue reverses the order
type ReverseByValue struct {
	ByKey
}

// Less checks the order of two element
func (l ReverseByValue) Less(i, j int) bool { return l.ByKey[i].Value > l.ByKey[j].Value }
