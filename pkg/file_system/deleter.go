package file_system

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bednarradek/php-deployer/pkg/ftp"
)

type Deleter interface {
	Delete(ctx context.Context, path string) error
	DeleteDir(ctx context.Context, path string) error
}

type SystemDeleter struct {
}

func NewSystemDeleter() *SystemDeleter {
	return &SystemDeleter{}
}

func (s SystemDeleter) Delete(_ context.Context, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("SystemDeleter::Delete error while deleting file %s: %w", path, err)
	}
	return nil
}

func (s SystemDeleter) DeleteDir(_ context.Context, path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("SystemDeleter::DeleteDir error while deleting directory %s: %w", path, err)
	}
	return nil
}

type FtpDeleter struct {
	ftpConnection *ftp.Connection
}

func NewFtpDeleter(ftpConnection *ftp.Connection) *FtpDeleter {
	return &FtpDeleter{ftpConnection: ftpConnection}
}

func (f FtpDeleter) Delete(ctx context.Context, path string) error {
	_, err := f.ftpConnection.FileSize(ctx, path)
	if err != nil {
		if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
			return nil
		}
		return err
	}
	if err := f.ftpConnection.Delete(ctx, path); err != nil {
		return err
	}
	return nil
}

func (f FtpDeleter) DeleteDir(ctx context.Context, path string) error {
	if err := f.ftpConnection.RemoveDirRecur(ctx, path); err != nil {
		if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
			return nil
		}
		return err
	}
	return nil
}
