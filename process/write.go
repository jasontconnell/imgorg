package process

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/jasontconnell/imgorg/data"
)

func Write(dst string, list []data.File, chunksize int) error {
	byhash := make(map[string][]data.File)
	for _, f := range list {
		byhash[f.Hash] = append(byhash[f.Hash], f)
	}

	grouped := [][]data.File{}
	for _, v := range byhash {
		grouped = append(grouped, v)
	}

	chunks := len(grouped) / chunksize
	if len(grouped)%chunksize != 0 {
		chunks++
	}

	var wg sync.WaitGroup
	wg.Add(chunks)
	for i := 0; i < chunks; i++ {
		start, end := i*chunksize, (i+1)*chunksize
		if end > len(grouped) {
			end = len(grouped)
		}

		go func(dst string, files [][]data.File) {
			for _, f := range files {
				handleFiles(dst, f)
			}
			wg.Done()
		}(dst, grouped[start:end])
	}

	wg.Wait()

	return nil
}

// handles files with the same hash
func handleFiles(dst string, files []data.File) {
	selectedFiles := pickFiles(files)
	if len(selectedFiles) == 0 {
		return
	}

	for _, selected := range selectedFiles {
		timedir := GetFolderFormat(selected.Mod)

		var root string = selected.Root
		if root == "" {
			root = timedir
		}

		var sub string = selected.Sub
		if sub == root {
			sub = ""
		}
		absdir := filepath.Join(dst, root, sub)

		srcPath := selected.Path
		destPath := filepath.Join(absdir, selected.Name)

		log.Println("copying from", srcPath, "to", destPath)

		_, err := os.Stat(absdir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(absdir, os.ModePerm)
			if err != nil {
				log.Println("couldn't make dir", absdir)
				return
			}
		} else if err != nil {
			log.Println("other err occurred ", err)
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

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			log.Println("couldn't copy contents "+selected.Path, "to", destPath, err)
			continue
		}

		log.Println("copied", selected.Path, "to", destPath, ". pretend deleting.")
		log.Println("os.Remove(", srcFile, ")")
	}
}

func pickFiles(files []data.File) []data.File {
	if len(files) <= 1 {
		return files
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Mod.After(files[j].Mod)
	})

	added := make(map[string]data.File)
	for _, f := range files {
		af, isAdded := added[f.Name]

		toAdd := f.Copy()
		if isAdded {
			toAdd.Name = fmt.Sprintf("%s_%s", f.Mod.Format(DateFormat), f.Name)
			toAdd.Root = af.Root // copy to same location
			toAdd.Sub = af.Sub
		}
		added[toAdd.Name] = toAdd
	}

	ret := []data.File{}
	for _, f := range added {
		ret = append(ret, f)
	}
	return ret
}
