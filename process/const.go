package process

import (
	"path/filepath"
	"time"
)

const DateFormat string = "20060102"

const DryRunCopyMessage string = `
===== Dry Run Copy File =====
  Source      :      %s
  Destination :      %s
  Roots       :      %v
  Mapped      :      %s
=============================`

const CopyMessage string = `
========   Copy Stats    =====
  Source      :      %s
  Destination :      %s
  Roots       :      %v
  Mapped      :      %s
  Success     :      %t
  File Size   :      %d
=============================`

// func GetFolderFormat(t time.Time) string {
// 	return filepath.Join(t.Format("2006"), t.Format("01"), t.Format("02"))
// }

func GetShortDateFormat(t time.Time) string {
	return filepath.Join(t.Format("2006"), t.Format("01"))
}
