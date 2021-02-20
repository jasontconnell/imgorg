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

func Read(job ImgOrgJob) ([]data.Dir, []data.File, int64, error) {
	list := []data.Dir{}

	for _, path := range job.Paths {
		filepath.Walk(path, func(f string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println("error from fp walk", f, err)
				return err
			}

			if !info.IsDir() {
				return nil
			}

			lname := strings.ToLower(info.Name())
			if _, ok := job.Ignore[lname]; ok {
				return filepath.SkipDir
			}

			var mapped string
			var rootlist []string

			lpath := strings.ToLower(f)
			parts := strings.Split(strings.TrimPrefix(lpath, filepath.VolumeName(f)), string(filepath.Separator))
			for _, p := range parts {
				if r, ok := job.Roots[p]; ok {
					rootlist = append(rootlist, r)
				}

				// get first mapped folder
				if s, ok := job.Mapped[p]; mapped == "" && ok {
					mapped = s
				}
			}

			d := data.Dir{Name: info.Name(), Path: f, Mapped: mapped, Roots: rootlist}
			list = append(list, d)

			return nil
		})
	}

	msgs := make(chan string, 1000000)

	files, written, err := readFiles(list, job, msgs)
	close(msgs)

	if job.Verbose {
		for msg := range msgs {
			log.Println(msg)
		}
	}

	return list, files, written, err
}

func readFiles(dirs []data.Dir, job ImgOrgJob, msgs chan string) ([]data.File, int64, error) {
	var wg sync.WaitGroup
	chdir := make(chan data.Dir, len(dirs))
	chfiles := make(chan *data.File, 100000)
	chhash := make(chan *data.File, 100000)

	for _, dir := range dirs {
		chdir <- dir
	}
	close(chdir)

	var size int64

	wg.Add(job.Workers)
	for i := 0; i < job.Workers; i++ {
		go func(chpaths chan data.Dir, exts map[string]string, chfile chan *data.File, chres chan *data.File) {
			done := false
			for !done {
				select {
				case dir := <-chpaths:
					done = len(chpaths) == 0 && len(chfile) == 0
					if dir.Path == "" {
						continue
					}

					entries, err := os.ReadDir(dir.Path)
					if err != nil {
						log.Println("couldn't read dir", dir.Path, len(chpaths), len(chfile))
						continue
					}

					if len(entries) == 0 {
						continue
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
							Name:   entry.Name(),
							Size:   info.Size(),
							Path:   filepath.Join(dir.Path, entry.Name()),
							Mod:    info.ModTime(),
							Roots:  dir.Roots,
							Mapped: dir.Mapped,
							Sub:    dir.Name,
						}

						atomic.AddInt64(&size, file.Size)

						if job.Verbose {
							msgs <- fmt.Sprintf("adding file %s with roots %v and mapped %s and sub %s", file.Path, file.Roots, file.Mapped, file.Sub)
						}

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
		}(chdir, job.Exts, chfiles, chhash)
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
	done <- true // kill the progress 'bar'
	close(chfiles)
	close(chhash)

	fileList := []data.File{}
	for f := range chhash {
		fileList = append(fileList, *f)
	}
	return fileList, size, nil
}

func hash(f *data.File) (string, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return "", err
	}
	return sha256sum(b), nil
}
