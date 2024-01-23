package file_system

import (
	"context"
	"errors"
	"fmt"

	"github.com/bednarradek/php-deployer/pkg/ftp"
)

type ChangeModer interface {
	Change(ctx context.Context, path string, mode string) error
}

type FtpChangeModer struct {
	ftpConnection *ftp.Connection
}

func NewFtpChangeModer(ftpConnection *ftp.Connection) *FtpChangeModer {
	return &FtpChangeModer{ftpConnection: ftpConnection}
}

func (f FtpChangeModer) Change(ctx context.Context, path string, mode string) error {
	if err := f.ftpConnection.Chmod(ctx, path, mode); err != nil {
		if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
			return nil
		}
		return fmt.Errorf("FtpChangeModer::Change error while changing mode of file %s: %w", path, err)
	}
	return nil
}
