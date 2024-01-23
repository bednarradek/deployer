package file_system

import (
	"context"
	"fmt"
)

type FileMover struct {
	reader Reader
	writer Writer
}

func NewFileMover(reader Reader, writer Writer) *FileMover {
	return &FileMover{reader: reader, writer: writer}
}

func (f FileMover) Move(ctx context.Context, pathFrom string, pathTo string) error {
	content, err := f.reader.Read(ctx, pathFrom)
	if err != nil {
		return fmt.Errorf("FileMover::Move error while reading file %s: %w", pathFrom, err)
	}
	if err := f.writer.Write(ctx, pathTo, content); err != nil {
		return fmt.Errorf("FileMover::Move error while writing file %s: %w", pathTo, err)
	}
	return nil
}
