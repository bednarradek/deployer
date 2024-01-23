package file_system

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"

	"github.com/bednarradek/php-deployer/pkg/ftp"
)

type Writer interface {
	Write(ctx context.Context, path string, data []byte) error
}

type SystemWriter struct {
}

func NewSystemWriter() *SystemWriter {
	return &SystemWriter{}
}

func (s SystemWriter) Write(_ context.Context, path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("SystemWriter::Write error while creating file %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("SystemWriter::Write error while writing file %s: %w", path, err)
	}
	return nil
}

type FtpWriter struct {
	ftpConnection *ftp.Connection
	defaultMode   string
}

func NewFtpWriter(ftpConnection *ftp.Connection, defaultMode string) *FtpWriter {
	return &FtpWriter{ftpConnection: ftpConnection, defaultMode: defaultMode}
}

func (f FtpWriter) Write(ctx context.Context, path string, data []byte) error {
	buff := bytes.NewBuffer(data)
	if err := f.ftpConnection.Stor(ctx, path, buff); err != nil {
		return fmt.Errorf("FtpWriter::Write error while writing file %s: %w", path, err)
	}
	if err := f.ftpConnection.Chmod(ctx, path, f.defaultMode); err != nil {
		return fmt.Errorf("FtpWriter::Write error while changing mode of file %s: %w", path, err)
	}
	return nil
}

type CompressionWriter struct {
	writer Writer
}

func NewCompressionWriter(writer Writer) *CompressionWriter {
	return &CompressionWriter{writer: writer}
}

func (c CompressionWriter) Write(ctx context.Context, path string, data []byte) error {
	compressed := new(bytes.Buffer)
	writer := gzip.NewWriter(compressed)
	defer func() {
		_ = writer.Close()
	}()
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("CompressionWriter::Write error while compressing data: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("CompressionWriter::Write error while flushing compressed data: %w", err)
	}
	if err := c.writer.Write(ctx, path, compressed.Bytes()); err != nil {
		return fmt.Errorf("CompressionWriter::Write error while uploading compressed data: %w", err)
	}
	return nil
}
