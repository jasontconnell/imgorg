package process

type ImgOrgJob struct {
	Paths   []string
	Roots   map[string]string
	Exts    map[string]string
	Mapped  map[string]string
	Ignore  map[string]string
	Workers int
	Delete  bool
	DryRun  bool
	Verbose bool
}
