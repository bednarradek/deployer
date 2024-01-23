package file_system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	ftp2 "github.com/bednarradek/php-deployer/pkg/ftp"
)

type Lister interface {
	List(ctx context.Context, dir string) ([]FileSystemObject, error)
}

type RecursiveLister struct {
	lister Lister
}

func NewRecursiveLister(lister Lister) *RecursiveLister {
	return &RecursiveLister{lister: lister}
}

func (r RecursiveLister) List(ctx context.Context, dir string) ([]FileSystemObject, error) {
	return r.list(ctx, dir, "")
}

func (r RecursiveLister) list(ctx context.Context, dir string, rel string) ([]FileSystemObject, error) {
	list, err := r.lister.List(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("RecursiveLister::list error while reading directory %s: %w", dir, err)
	}
	res := make([]FileSystemObject, 0, len(list))
	for _, file := range list {
		absPath := fmt.Sprintf("%s/%s", dir, file.GetName())
		relPath := fmt.Sprintf("%s/%s", rel, file.GetName())
		if file.IsDir() {
			res = append(res, NewRecursiveObject(relPath, true))
			r, err := r.list(ctx, absPath, relPath)
			if err != nil {
				return nil, fmt.Errorf("RecursiveLister::list error while reading directory %s: %w", dir, err)
			}
			res = append(res, r...)
			continue
		}
		if file.IsRegular() {
			res = append(res, NewRecursiveObject(relPath, false))
			continue
		}
	}
	return res, nil
}

type SystemLister struct {
}

func NewSystemLister() *SystemLister {
	return &SystemLister{}
}

func (s SystemLister) List(_ context.Context, dir string) ([]FileSystemObject, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("SystemLoader::list error while reading directory %s: %w", dir, err)
	}
	res := make([]FileSystemObject, 0, len(list))
	for _, file := range list {
		res = append(res, SystemObject{file})
	}
	return res, nil
}

type FtpLister struct {
	ftpConnection *ftp2.Connection
}

func NewFtpLister(ftpConnection *ftp2.Connection) *FtpLister {
	return &FtpLister{ftpConnection: ftpConnection}
}

func (f FtpLister) List(ctx context.Context, dir string) ([]FileSystemObject, error) {
	list, err := f.ftpConnection.List(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("FtpLister::List error while reading directory %s: %w", dir, err)
	}
	res := make([]FileSystemObject, 0, len(list))
	for _, file := range list {
		res = append(res, FtpObject{*file})
	}
	return res, err
}

type LogLister struct {
	path   string
	lister Lister
	reader Reader
}

func NewLogLister(path string, lister Lister, reader Reader) *LogLister {
	return &LogLister{path: path, lister: lister, reader: reader}
}

func (l *LogLister) GetLogFile(ctx context.Context) (*LogFile, error) {
	result := new(LogFile)
	logContent, err := l.reader.Read(ctx, l.path)
	if err != nil {
		return nil, fmt.Errorf("LogLister::GetLogFile error while reading log file %s: %w", l.path, err)
	}
	if logContent == nil {
		return nil, nil
	}
	if err := json.Unmarshal(logContent, result); err != nil {
		return nil, fmt.Errorf("LogLister::GetLogFile error while unmarshalling log file %s: %w", l.path, err)
	}
	return result, nil
}

func (l *LogLister) List(ctx context.Context, dir string) ([]FileSystemObject, error) {
	logFile, err := l.GetLogFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("LogLister::List error while getting log file: %w", err)
	}
	// use backup lister when log file is missing
	if logFile == nil {
		return l.lister.List(ctx, dir)
	}
	result := make([]FileSystemObject, 0, 100)
	for _, object := range logFile.Objects {
		result = append(result, object)
	}
	return result, nil
}
