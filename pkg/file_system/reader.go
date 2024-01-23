package file_system

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bednarradek/php-deployer/pkg/ftp"
)

type Reader interface {
	Read(ctx context.Context, path string) ([]byte, error)
}

type SystemReader struct {
}

func NewSystemReader() *SystemReader {
	return &SystemReader{}
}

func (s SystemReader) Read(_ context.Context, path string) ([]byte, error) {
	res, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("SystemReader::Read error while reading file %s: %w", path, err)
	}
	return res, nil
}

type FtpReader struct {
	ftpConnection *ftp.Connection
}

func NewFtpReader(ftpConnection *ftp.Connection) *FtpReader {
	return &FtpReader{ftpConnection: ftpConnection}
}

func (f *FtpReader) Read(ctx context.Context, path string) ([]byte, error) {
	response, err := f.ftpConnection.Retr(ctx, path)
	if err != nil {
		if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
			return nil, nil
		}
		return nil, fmt.Errorf("FtpReader::Read error while reading file %s: %w", path, err)
	}
	defer func() {
		_ = response.Close()
	}()
	res, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("FtpReader::Read error while reading file %s: %w", path, err)
	}
	return res, nil
}

type CompressionReader struct {
	reader Reader
}

func NewCompressionReader(reader Reader) *CompressionReader {
	return &CompressionReader{reader: reader}
}

func (c CompressionReader) Read(ctx context.Context, path string) ([]byte, error) {
	b, err := c.reader.Read(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("CompressionReader::Read error while reading file %s: %w", path, err)
	}
	if b == nil {
		return nil, nil
	}
	read, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("CompressionReader::Read error while decompressing file %s: %w", path, err)
	}
	defer func() {
		_ = read.Close()
	}()
	data, err := io.ReadAll(read)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("CompressionReader::Read error while reading file %s: %w", path, err)
	}
	return data, nil
}
