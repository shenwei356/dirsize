// Copyright 2013 Wei Shen (shenwei356@gmail.com). All rights reserved.
// Use of this source code is governed by a MIT-license
// that can be found in the LICENSE file.

// Summarize size of directories and files in directories. 
// Version 3: speed up by goroutine.
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
	"sync"
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
			size, info, err := DirSize(arg, n)
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

func DirSize(path string, n int) (float64, []Item, error) {
	var size float64 = 0
	var info []Item
	info = make([]Item, 0)

	bytes, err := ioutil.ReadFile(path)
	if err == nil {
		size1 := float64(len(bytes))
		info = append(info, Item{path, size1})
		return size1, info, nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		recover()
		return 0, nil, errors.New("ReadDir Error: " + path)
	}

	sizes := make(map[string]float64, 0)
	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		sizes[file.Name()] = 0
	}

	// walk
	var wg sync.WaitGroup
	tokens := make(chan int, n)
	err = filepath.Walk(path, func(file string, fileinfo os.FileInfo, err error) error {
		if file == path {
			return nil
		}
		tokens <- 1
		wg.Add(1)
		go func() {
			for k, _ := range sizes {
				if strings.HasPrefix(file, filepath.Join(path, k)+string(filepath.Separator)) || file == filepath.Join(path, k) {
					sizes[k] += float64(fileinfo.Size())
					break
				}
			}
			wg.Done()
			<-tokens
		}()
		return nil
	})
	if err != nil {
		return 0, nil, errors.New("Walk Dir Error: " + path)
	}
	wg.Wait()

	for k, v := range sizes {
		info = append(info, Item{k, v})
		size += v
	}
	return size, info, nil
}
