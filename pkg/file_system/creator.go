package file_system

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bednarradek/php-deployer/pkg/ftp"
)

type Creator interface {
	CreateDir(ctx context.Context, path string) error
}

type FtpCreator struct {
	ftpConnection *ftp.Connection
	defaultMode   string
}

func NewFtpCreator(ftpConnection *ftp.Connection, defaultMode string) *FtpCreator {
	return &FtpCreator{ftpConnection: ftpConnection, defaultMode: defaultMode}
}

func (f FtpCreator) CreateDir(ctx context.Context, path string) error {
	split := strings.Split(path, "/")
	for i := 0; i < len(split); i++ {
		p := strings.Join(split[:i+1], "/")
		if err := f.ftpConnection.MakeDir(ctx, p); err != nil {
			if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
				continue
			}
			return fmt.Errorf("FtpCreator::CreateDir error while creating directory %s from %s: %w", p, path, err)
		}
		if err := f.ftpConnection.Chmod(ctx, p, f.defaultMode); err != nil {
			if errors.Is(err, ftp.ErrorFtpPermissionDenied) {
				continue
			}
			return fmt.Errorf("FtpCreator::CreateDir error while changing mode of directory %s from %s: %w", p, path, err)
		}
	}
	return nil
}
