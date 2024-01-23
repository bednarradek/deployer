package helpers

import "strings"

func GetDirectoryPath(path string) string {
	slashIndex := strings.LastIndex(path, "/")
	dirs := path[:slashIndex]
	return dirs
}
