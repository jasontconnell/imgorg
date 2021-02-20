package data

import (
	"time"
)

type Dir struct {
	Name   string
	Path   string
	Mapped string
	Parts  []string
	Roots  []string
}

type File struct {
	Name   string
	Path   string
	Sub    string
	Hash   string
	Mod    time.Time
	Size   int64
	Roots  []string
	Mapped string
}

func (f File) Copy() File {
	rc := make([]string, len(f.Roots))
	copy(rc, f.Roots)
	nf := File{
		Name:  f.Name,
		Path:  f.Path,
		Sub:   f.Sub,
		Hash:  f.Hash,
		Mod:   f.Mod,
		Size:  f.Size,
		Roots: rc,
	}
	return nf
}
