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

const imageExtensions string = "jpg,jpeg,gif,bmp,png,psd,tif,svg,tga"

func main() {
	base := flag.String("base", "", "source folder")
	subs := flag.String("sub", "", "optional csv list of subfolders. Only these will be scanned")
	dst := flag.String("dst", "", "destination folder")
	exts := flag.String("exts", "", "image extensions")
	roots := flag.String("roots", "", "csv of root organization folders")
	ignores := flag.String("ignore", "", "csv of folders to ignore")
	fmap := flag.String("map", "", "map folders to other folders")
	workers := flag.Int("workers", 20, "number of workers")
	dryrun := flag.Bool("dryrun", false, "just output the files and their new locations")
	delsrc := flag.Bool("delete", false, "specify this to also deleted the source files after copying")
	verbose := flag.Bool("verbose", false, "specify this to see every move imgorg makes")
	flag.Parse()

	start := time.Now()
	log.SetOutput(os.Stdout)

	if *dryrun {
		log.Println("**************** dry run ****************")
	}

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

	fm := make(map[string]string)
	for _, s := range strings.Split(*fmap, ",") {
		vs := strings.Split(s, "=")
		if len(vs) == 2 {
			k, v := strings.ToLower(vs[0]), vs[1]
			fm[k] = v
		}
	}

	job := process.ImgOrgJob{
		Paths:   paths,
		Roots:   rmap,
		Mapped:  fm,
		Exts:    extmap,
		Ignore:  imap,
		Workers: *workers,
		DryRun:  *dryrun,
		Delete:  *delsrc,
		Verbose: *verbose,
	}

	dirs, files, read, err := process.Read(job)
	if err != nil {
		log.Fatal(err)
	}

	written, err := process.Write(*dst, files, job)
	if err != nil {
		log.Fatal(err)
	}

	if job.Delete {
		process.Delete(files, job)
		process.ClearEmptyDirectories(dirs, job)
	}

	log.Println("finished. read:", read, "wrote:", written, time.Since(start))
}

func mapStrings(strs []string) map[string]string {
	m := make(map[string]string)
	for _, s := range strs {
		m[strings.ToLower(s)] = s
	}
	return m
}
