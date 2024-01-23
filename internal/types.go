package internal

type Deployer interface {
	Deploy() error
}

type CompareResult struct {
	Object CompareObject
	Action string
}

type CompareObject interface {
	Path() string
	IsDir() bool
	Hash() string
	GetKey() string
}

type File struct {
	path string
	hash string
}

func NewFile(path string, hash string) *File {
	return &File{path: path, hash: hash}
}

func (f File) Path() string {
	return f.path
}

func (f File) IsDir() bool {
	return false
}

func (f File) Hash() string {
	return f.hash
}

func (f File) GetKey() string {
	return f.path
}

type Folder struct {
	path string
}

func NewFolder(path string) *Folder {
	return &Folder{path: path}
}

func (f Folder) Path() string {
	return f.path
}

func (f Folder) IsDir() bool {
	return true
}

func (f Folder) Hash() string {
	return ""
}

func (f Folder) GetKey() string {
	return f.path
}
