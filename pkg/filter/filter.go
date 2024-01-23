package filter

import (
	"regexp"
)

type Filter interface {
	Contain(path string) bool
}

type FileSystemFilter struct {
	ignoreList []string
}

func NewFileSystemFilter(ignoreList []string) *FileSystemFilter {
	return &FileSystemFilter{ignoreList: ignoreList}
}

func (f *FileSystemFilter) Contain(path string) bool {
	for _, ignore := range f.ignoreList {
		match, _ := regexp.MatchString(ignore, path)
		if match {
			return true
		}
	}
	return false
}
