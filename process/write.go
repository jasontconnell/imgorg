package process

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/jasontconnell/imgorg/data"
)

func Write(dst string, list []data.File, job ImgOrgJob) (int64, error) {
	byhash := make(map[string][]data.File)
	for _, f := range list {
		byhash[f.Hash] = append(byhash[f.Hash], f)
	}

	grouped := [][]data.File{}
	for _, v := range byhash {
		grouped = append(grouped, v)
	}

	chunks := len(grouped) / job.Workers
	if len(grouped)%job.Workers != 0 {
		chunks++
	}

	var written int64
	msgs := make(chan string, 1000000)
	var wg sync.WaitGroup
	wg.Add(chunks)
	for i := 0; i < chunks; i++ {
		start, end := i*job.Workers, (i+1)*job.Workers
		if end > len(grouped) {
			end = len(grouped)
		}

		go func(dst string, files [][]data.File) {
			for _, f := range files {
				w := handleFiles(dst, f, job, msgs)
				atomic.AddInt64(&written, w)
			}
			wg.Done()
		}(dst, grouped[start:end])
	}

	wg.Wait()
	close(msgs)

	if job.Verbose || job.DryRun {
		for msg := range msgs {
			fmt.Println(msg)
		}
	}

	return written, nil
}

// handles files with the same hash
func handleFiles(dst string, files []data.File, job ImgOrgJob, msgs chan string) int64 {
	selectedFiles := pickFiles(files, job, msgs)
	if len(selectedFiles) == 0 {
		return 0
	}
	var bytesWritten int64

	for _, selected := range selectedFiles {
		timedir := GetShortDateFormat(selected.Mod)

		var roots []string = selected.Roots
		hasRoots := true
		if len(roots) == 0 {
			roots = []string{timedir}
			hasRoots = false
		}

		absdir := filepath.Join(dst)
		if hasRoots {
			for _, r := range roots {
				absdir = filepath.Join(absdir, r)
			}
			absdir = filepath.Join(absdir, selected.Sub)
		} else {
			// if it doesn't have roots or is mapped, mapped will be blank
			// and filepath.Join can work with a blank.
			absdir = filepath.Join(absdir, selected.Mapped, timedir)
		}

		srcPath := selected.Path
		destPath := filepath.Join(absdir, selected.Name)

		if job.DryRun {
			msg := fmt.Sprintf(DryRunCopyMessage, srcPath, destPath, selected.Roots, selected.Mapped)
			msgs <- msg
			continue
		}

		// check if file exists, if not, create a hash based folder for it
		_, err := os.Stat(destPath)
		if err == nil {
			absdir = filepath.Join(absdir, string(selected.Hash[:10]))
			destPath = filepath.Join(absdir, selected.Name)
		}

		_, err = os.Stat(absdir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(absdir, os.ModePerm)
			if err != nil {
				log.Println("couldn't make dir", absdir)
				return 0
			}
		} else if err != nil {
			log.Println("other err occurred ", absdir, err)
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			log.Println("couldn't create file "+destPath, err)
			continue
		}
		defer destFile.Close()

		srcFile, err := os.Open(srcPath)
		if err != nil {
			log.Println("couldn't open file to copy "+srcPath, err)
			continue
		}
		defer srcFile.Close()

		w, err := io.Copy(destFile, srcFile)
		if err != nil {
			log.Println("couldn't copy contents", srcPath, "to", destPath, err)
			continue
		}

		bytesWritten += w

		err = os.Chtimes(destPath, selected.Mod, selected.Mod)
		if err != nil {
			log.Println("couldn't update modified time for", destPath)
			continue
		}

		if job.Verbose {
			msg := fmt.Sprintf(CopyMessage, srcPath, destPath, selected.Roots, selected.Mapped, true, w)
			msgs <- msg
		}
	}
	return bytesWritten
}

func pickFiles(files []data.File, job ImgOrgJob, msgs chan string) []data.File {
	if len(files) <= 1 {
		return files
	}

	sort.Slice(files, func(i, j int) bool {
		mappedi := files[i].Mapped
		mappedj := files[j].Mapped

		rootsi := files[i].Roots
		rootsj := files[j].Roots

		mapsort := false
		rootsort := false
		isLess := false
		if len(rootsi) > len(rootsj) {
			isLess = true
			rootsort = true
		}

		if !rootsort && mappedi != mappedj {
			mapsort = true
			isLess = mappedi != "" && mappedj == ""
		}

		if !mapsort {
			isLess = files[i].Mod.After(files[j].Mod)
		}

		return isLess
	})

	added := make(map[string]data.File)
	for _, f := range files {
		af, isAdded := added[f.Name]

		if job.Verbose {
			msgs <- fmt.Sprintf("handling file %s. Is added is %t by file in path %s", f.Path, isAdded, af.Path)
		}

		// hash is the same or it wouldn't be in this list
		// if the mod time is the same it was copied to another location
		// and should be skipped, we can assume it's a copy.
		if f.Mod.Equal(af.Mod) {
			if job.Verbose {
				msgs <- fmt.Sprintf("file %s is the same mod time as already added %s", f.Path, af.Path)
			}
			continue
		}

		toAdd := f.Copy()
		if isAdded {
			toAdd.Name = fmt.Sprintf("%s_%s", f.Mod.Format(DateFormat), f.Name)
			// copy to same location
			toAdd.Roots = af.Roots
			toAdd.Mapped = af.Mapped
			toAdd.Sub = af.Sub

			if job.Verbose {
				msgs <- fmt.Sprintf("file %s copying attributes from %s\nroots: %v\nmapped: %s\nsub:%s", f.Path, af.Path, af.Roots, af.Mapped, af.Sub)
			}
		}

		if job.Verbose {
			msgs <- fmt.Sprintf("adding file %s with attributes\nroots: %v\nmapped: %s\nsub:%s", toAdd.Path, toAdd.Roots, toAdd.Mapped, toAdd.Sub)
		}
		added[toAdd.Name] = toAdd
	}

	ret := []data.File{}
	for _, f := range added {
		ret = append(ret, f)
	}
	return ret
}
