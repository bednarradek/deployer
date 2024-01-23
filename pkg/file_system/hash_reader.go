package file_system

import (
	"context"
	"fmt"

	"github.com/bednarradek/php-deployer/pkg/helpers"
)

type HashReader interface {
	ReadHash(ctx context.Context, path string) (string, error)
}

type StandardHashReader struct {
	reader Reader
}

func NewStandardHashReader(reader Reader) *StandardHashReader {
	return &StandardHashReader{reader: reader}
}

func (s *StandardHashReader) ReadHash(ctx context.Context, path string) (string, error) {
	content, err := s.reader.Read(ctx, path)
	if err != nil {
		return "", fmt.Errorf("StandardHashReader::ReadHash error while reading file %s: %w", path, err)
	}
	return helpers.HashBytes(content), nil
}

type LogHashReader struct {
	logFile LogFile
	objects map[string]LogObject
	reader  HashReader
}

func NewLogHashReader(logFile LogFile, reader HashReader) *LogHashReader {
	objects := helpers.ConvertToMap(logFile.Objects)
	return &LogHashReader{
		logFile: logFile,
		objects: objects,
		reader:  reader,
	}
}

func (l *LogHashReader) ReadHash(ctx context.Context, path string) (string, error) {
	if o, ok := l.objects[path]; ok {
		return o.Hash, nil
	}
	return l.reader.ReadHash(ctx, path)
}
