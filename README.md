dirsize
========

Command tool for summarizing size of directories and files in directories.
dirsize is wiritten in [golang](http://golang.org).

Install
-------
This package is "go-gettable", just:

    go get github.com/shenwei356/util/dirsize

Usage
-----
    
    Usage: dirsize [OPTION]... [DIR]...

    Summarize size of directories and files in directories.

    OPTION:
      -a=false: sort by Alphabet.
      -r=false: sort by Size.
      -s=true: reverse order while sorting.

Example
-------
    
    dirsize -s -r .

Result:

    .:   2.40 MB
      0.00  B	t.txt
    483.00  B	.gitattributes
      1.07 KB	LICENSE
      1.15 KB	README.md
      1.92 KB	.gitignore
      2.76 KB	main.go
      1.02 MB	.git
      1.37 MB	dirsize.exe

Have a Try
----------
You can compile by yourself or just download the executable files immediately.

- [dirsize.exe](https://github.com/shenwei356/dirsize/blob/master/dirsize.exe) for win32.

Attention
---------
The ouput text is encoded in UTF-8. For those who use none-utf8 shell, 
redirecting to a temporary file is an easy solution.

    dirsize -s -r . > tmp.txt
      
Copyright (c) 2013, Wei Shen (shenwei356@gmail.com)

[MIT License](https://github.com/shenwei356/dirsize/blob/master/LICENSE)