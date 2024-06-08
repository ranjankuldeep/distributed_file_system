package store

import (
	"fmt"
	"strings"
)

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// Returns the full path excluding the root folder with the filename.
// Eg- dslkjfdsf/jdfsfj/fdsf/dnf/<filename>
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}
