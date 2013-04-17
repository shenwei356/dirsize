// Copyright 2013 Wei Shen (shenwei356@gmail.com). All rights reserved.
// Use of this source code is governed by a MIT-license
// that can be found in the LICENSE file.

// Summarize size of directories and files in directories. 
// Version 2: speed up by goroutine.
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
	"runtime"
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
		fmt.Println("Version 2: speed up by goroutine.")
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
	n := runtime.NumCPU()
	if n > 2 {
		n = n - 1
	} else {
		n = 2
	}
	runtime.GOMAXPROCS(n)

	for _, arg := range flag.Args() {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			msg := make(chan DirSizeInfo, 1)
			FolderSize(arg, msg, true)
			m := <-msg

			if m.Err != nil {
				fmt.Println(m.Err)
			}
			info := m.Info
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

			fmt.Printf("\n%s: %v\n", arg, ByteSize(m.Size))
			for _, item := range info {
				fmt.Printf("%v\t%s\n", ByteSize(item.Value), item.Key)
			}

		} else {
			fmt.Printf("\n%s: NOT exists!\n", arg)
		}
	}
}

type DirSizeInfo struct {
	Name string
	Size float64
	Info []Item
	Err  error
}

// Get total size of files in a directory, and store the sizes of first level
// directories and files in a key-value list.
func FolderSize(dirname string, msg chan DirSizeInfo, firstLevel bool) {
	var size float64 = 0
	var info []Item
	if firstLevel {
		info = make([]Item, 0)
	}

	// dirname is a file
	bytes, err := ioutil.ReadFile(dirname)
	if err == nil {
		size1 := float64(len(bytes))
		if firstLevel {
			info = append(info, Item{dirname, size1})
		}
		msg <- DirSizeInfo{dirname, size1, info, nil}
		return
	}

	// ReadDir Error!
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		recover()
		msg <- DirSizeInfo{dirname, 0, nil, errors.New("ReadDir Error: " + dirname)}
		return
	}

	// read directories
	dirs := make([]os.FileInfo, 0)
	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		if file.IsDir() { // dir
			dirs = append(dirs, file) // commpute size later
		} else { // file
			size1 := float64(file.Size())
			size += size1
			if firstLevel {
				info = append(info, Item{file.Name(), size1})
			}
		}
	}

	// sub directories
	n := len(dirs)
	c := make(chan DirSizeInfo, n)
	for _, dir := range dirs {
		if firstLevel {
			go FolderSize(filepath.Join(dirname, dir.Name()), c, false)
		} else { // avoid creating too many goroutines.
			FolderSize(filepath.Join(dirname, dir.Name()), c, false)
		}
	}

	for i := 0; i < n; i++ {
		m := <-c
		if m.Err != nil {
			msg <- DirSizeInfo{dirname, 0, nil, m.Err}
			return
		}
		size += m.Size
		if firstLevel {
			info = append(info, Item{m.Name, m.Size})
		}
	}
	msg <- DirSizeInfo{dirname, size, info, nil}
	return
}
