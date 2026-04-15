package util

import (
	"net/url"
	"path/filepath"
)

func PathJoin(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}
func IsURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme != "" && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp" || u.Scheme == "ftps" || u.Scheme == "ws" || u.Scheme == "wss")
}

func IsFilePath(s string) bool {
	if len(s) > 1 && (s[0] == '/' || s[1] == ':') {
		return true
	}
	if len(s) > 2 && s[0:2] == "./" || s[0:2] == ".." {
		return true
	}
	u, err := url.Parse(s)
	if err == nil && u.Scheme == "file" {
		return true
	}
	return false
}
