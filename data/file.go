package data

import "time"

type File struct {
	Name string
	Path string
	Hash string
	Mod  time.Time
	Size int64
	Root string
	Sub  string
}

func (f File) Copy() File {
	nf := File{
		Name: f.Name,
		Path: f.Path,
		Hash: f.Hash,
		Mod:  f.Mod,
		Size: f.Size,
		Root: f.Root,
		Sub:  f.Sub,
	}
	return nf
}
