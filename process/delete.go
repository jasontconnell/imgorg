package process

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jasontconnell/imgorg/data"
)

func Delete(list []data.File, job ImgOrgJob) {
	for _, f := range list {
		if job.DryRun {
			log.Println("Deleting (fake)", f.Path)
			continue
		}
		err := os.Remove(f.Path)
		if err != nil {
			log.Println("error deleting", f.Path, err.Error())
		}
	}
}

func ClearEmptyDirectories(list []data.Dir, job ImgOrgJob) {

	list = sortDirs(list)

	for _, dir := range list {
		entries, err := os.ReadDir(dir.Path)
		if err != nil {
			log.Println("couldn't get entries for", dir, err)
		}

		if len(entries) == 0 {
			if job.DryRun {
				log.Println("Removing (fake)", dir.Path)
				continue
			}
			err = os.Remove(dir.Path)
			if err != nil {
				log.Println("couldn't remove directory", dir, err)
			}
		}
	}
}

type sortdir struct {
	dir   data.Dir
	depth int
}

func sortDirs(list []data.Dir) []data.Dir {
	sorts := []sortdir{}
	// quick late night way to sort directories so they delete if they're empty
	for _, dir := range list {
		depth := strings.Count(dir.Path, string(filepath.Separator))
		sd := sortdir{dir, depth}
		sorts = append(sorts, sd)
	}

	sort.Slice(sorts, func(i, j int) bool {
		return sorts[i].depth > sorts[j].depth
	})

	newdirs := []data.Dir{}
	for _, srt := range sorts {
		newdirs = append(newdirs, srt.dir)
	}
	return newdirs
}
