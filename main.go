package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jasontconnell/imgorg/process"
)

const imageExtensions string = "jpg,jpeg,gif,bmp,png,psd,tif,svg"

func main() {
	base := flag.String("base", "", "source folder")
	subs := flag.String("sub", "", "optional csv list of subfolders. Only these will be scanned")
	dst := flag.String("dst", "", "destination folder")
	exts := flag.String("exts", "", "image extensions")
	roots := flag.String("roots", "", "csv of root organization folders")
	ignores := flag.String("ignore", "", "csv of folders to ignore")
	workers := flag.Int("workers", 3, "number of workers")
	flag.Parse()

	start := time.Now()
	log.SetOutput(os.Stdout)

	if *exts == "" {
		*exts = imageExtensions
	}

	rmap := mapStrings(strings.Split(*roots, ","))
	imap := mapStrings(strings.Split(*ignores, ","))

	extmap := make(map[string]string)
	for _, ext := range strings.Split(*exts, ",") {
		lower := "." + strings.ToLower(ext)
		extmap[lower] = lower
	}

	if *base == "" || *dst == "" {
		flag.PrintDefaults()
		return
	}

	var paths []string
	if *subs != "" {
		subdirs := strings.Split(*subs, ",")
		for _, subdir := range subdirs {
			paths = append(paths, filepath.Join(*base, subdir))
		}
	} else {
		paths = []string{*base}
	}

	files, err := process.Read(paths, rmap, extmap, imap, *workers)
	if err != nil {
		log.Fatal(err)
	}

	err = process.Write(*dst, files, *workers)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("finished.", time.Since(start))
}

func mapStrings(strs []string) map[string]string {
	m := make(map[string]string)
	for _, s := range strs {
		m[s] = s
	}
	return m
}
