package internal

import (
	"context"
	"fmt"

	"github.com/bednarradek/php-deployer/pkg/file_system"
	"github.com/bednarradek/php-deployer/pkg/filter"
	"github.com/bednarradek/php-deployer/pkg/helpers"
)

type ReadManager struct {
	lister     file_system.Lister
	hashReader file_system.HashReader
	filter     filter.Filter
}

func NewReaderManager(lister file_system.Lister, hashReader file_system.HashReader, filter filter.Filter) *ReadManager {
	return &ReadManager{
		lister:     lister,
		hashReader: hashReader,
		filter:     filter,
	}
}

func (m *ReadManager) Read(ctx context.Context, dir string) ([]CompareObject, error) {
	files, err := m.lister.List(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("ReadManager::Read error while reading directory: %w", err)
	}

	result, err := helpers.RunWorkers(ctx, 10, files, func(ctx context.Context, input file_system.FileSystemObject) (CompareObject, error) {
		absDir := fmt.Sprintf("%s%s", dir, input.GetName())
		relDir := input.GetName()
		if m.filter.Contain(relDir) {
			return nil, nil
		}
		if input.IsDir() {
			return NewFolder(relDir), nil
		} else if input.IsRegular() {
			r, err := m.readFile(ctx, absDir, relDir)
			if err != nil {
				return nil, fmt.Errorf("ReadManager::readFile error while reading file %s: %w", absDir, err)
			}
			return r, nil
		}
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *ReadManager) readFile(ctx context.Context, path string, relativePath string) (CompareObject, error) {
	hash, err := m.hashReader.ReadHash(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("ReadManager::readFile error while reading hash for file %s: %w", path, err)
	}
	return NewFile(relativePath, hash), nil
}
