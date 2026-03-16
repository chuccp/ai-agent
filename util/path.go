package util

import "path/filepath"

func PathJoin(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}
