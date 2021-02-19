package process

import (
	"path/filepath"
	"time"
)

const DateFormat string = "20060102"

func GetFolderFormat(t time.Time) string {
	return filepath.Join(t.Format("2006"), t.Format("01"), t.Format("02"))
}
