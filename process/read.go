package process

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jasontconnell/imgorg/data"
)

func sha256sum(d []byte) string {
	h := sha256.New()
	h.Write(d)
	b := h.Sum(nil)

	return fmt.Sprintf("%x", b)
}

func Read(paths []string, exts map[string]string, roots, ignores map[string]string, chunksize int) ([]data.File, error) {
	list := []string{}

	for _, path := range paths {
		filepath.Walk(path, func(f string, info os.FileInfo, err error) error {
			if _, ok := ignores[info.Name()]; info.IsDir() && !ok {
				list = append(list, f)
			}
			return nil
		})
	}

	log.Println("len list", len(list))
	return readFiles(list, exts, roots, chunksize)
}

func readFiles(dirs []string, roots, exts map[string]string, workers int) ([]data.File, error) {

	var wg sync.WaitGroup
	chdir := make(chan string, len(dirs))
	chfiles := make(chan *data.File, 100000)
	chhash := make(chan *data.File, 100000)

	for _, dir := range dirs {
		chdir <- dir
	}
	close(chdir)

	var size int64

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(chpaths chan string, roots map[string]string, chfile chan *data.File, chres chan *data.File) {
			done := false
			for !done {
				select {
				case dir := <-chpaths:
					done = len(chpaths) == 0 && len(chfile) == 0
					if dir == "" {
						continue
					}
					entries, err := os.ReadDir(dir)
					if err != nil {
						log.Println("couldn't read dir", dir, len(chpaths), len(chfile))
						continue
					}

					if len(entries) == 0 {
						continue
					}

					var root string
					parts := strings.Split(strings.TrimPrefix(dir, filepath.VolumeName(dir)), string(filepath.Separator))
					for _, p := range parts {
						plower := strings.ToLower(p)
						if _, ok := roots[plower]; ok {
							root = p
							break
						}
					}

					var sub string
					if len(parts) > 1 {
						sub = parts[len(parts)-1]
					}

					for _, entry := range entries {
						if entry.IsDir() {
							continue
						}
						ext := strings.ToLower(filepath.Ext(entry.Name()))
						if _, ok := exts[ext]; !ok {
							continue
						}

						info, err := entry.Info()
						if err != nil {
							log.Println("couldn't stat file ", entry.Name(), err)
							continue
						}
						file := &data.File{
							Name: entry.Name(),
							Size: info.Size(),
							Path: filepath.Join(dir, entry.Name()),
							Mod:  info.ModTime(),
							Root: root,
							Sub:  sub,
						}

						atomic.AddInt64(&size, file.Size)

						chfile <- file
					}
				case f := <-chfile:
					done = len(chpaths) == 0 && len(chfile) == 0
					s, err := hash(f)
					if err != nil {
						log.Println("hashing file", f.Path, err)
						continue
					}
					f.Hash = s
					chres <- f
				default:
					done = len(chpaths) == 0 && len(chfile) == 0
					break
				}
			}

			wg.Done()
		}(chdir, roots, chfiles, chhash)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-time.Tick(2 * time.Second):
				fmt.Printf("\r\tdirectories left: %d file queue: %d files hashed: %d bytes read: %d", len(chdir), len(chfiles), len(chhash), size)
			case <-done:
				fmt.Println()
				return
			}
		}
	}()

	wg.Wait()
	done <- true
	close(chfiles)
	close(chhash)

	fileList := []data.File{}
	for f := range chhash {
		fileList = append(fileList, *f)
	}
	return fileList, nil
}

func hash(f *data.File) (string, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return "", err
	}
	return sha256sum(b), nil
}
