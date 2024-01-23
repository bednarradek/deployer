package file_system

import (
	"os"

	"github.com/bednarradek/ftp"
)

type FileSystemObject interface {
	IsDir() bool
	IsRegular() bool
	GetName() string
}

type SystemObject struct {
	os.DirEntry
}

func (s SystemObject) IsDir() bool {
	return s.Type().IsDir()
}

func (s SystemObject) IsRegular() bool {
	return s.Type().IsRegular()
}

func (s SystemObject) GetName() string {
	return s.Name()
}

type FtpObject struct {
	ftp.Entry
}

func (f FtpObject) IsDir() bool {
	return f.Type == ftp.EntryTypeFolder
}

func (f FtpObject) IsRegular() bool {
	return f.Type == ftp.EntryTypeFile
}

func (f FtpObject) GetName() string {
	return f.Name
}

type LogFile struct {
	Objects []LogObject `json:"objects"`
}

type LogObject struct {
	Path          string `json:"path"`
	IsDirFlag     bool   `json:"isDir"`
	IsRegularFlag bool   `json:"isRegular"`
	Hash          string `json:"hash"`
}

func (l LogObject) IsDir() bool {
	return l.IsDirFlag
}

func (l LogObject) IsRegular() bool {
	return l.IsRegularFlag
}

func (l LogObject) GetName() string {
	return l.Path
}

func (l LogObject) GetKey() string {
	return l.Path
}

type RecursiveObject struct {
	path string
	dir  bool
}

func NewRecursiveObject(path string, dir bool) *RecursiveObject {
	return &RecursiveObject{path: path, dir: dir}
}

func (r RecursiveObject) IsDir() bool {
	return r.dir
}

func (r RecursiveObject) IsRegular() bool {
	return !r.dir
}

func (r RecursiveObject) GetName() string {
	return r.path
}
