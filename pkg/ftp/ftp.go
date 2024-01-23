package ftp

import (
	"context"
	"fmt"
	"io"
	"net/textproto"
	"sync"
	"time"

	"github.com/bednarradek/ftp"
	"github.com/silenceper/pool"
)

const ftpPermissionDeniedCode = 550

var ErrorFtpPermissionDenied = fmt.Errorf("permission denied")

type Connection struct {
	once sync.Once

	url       string
	user      string
	password  string
	ftpConfig *pool.Config
	ftpPool   pool.Pool
}

func NewConnection(url string, user string, password string) *Connection {
	return &Connection{url: url, user: user, password: password}
}

func (f *Connection) Connect() (err error) {
	f.once.Do(func() {
		f.ftpConfig = &pool.Config{
			InitialCap: 5,
			MaxCap:     30,
			MaxIdle:    20,
			Factory: func() (interface{}, error) {
				ftpCon, err := ftp.Dial(f.url, ftp.DialWithTimeout(60*time.Second))
				if err != nil {
					return nil, fmt.Errorf("Connection::Connect error while connecting to FTP server: %w", err)
				}
				if err := ftpCon.Login(f.user, f.password); err != nil {
					return nil, fmt.Errorf("Connection::Connect error while logging to FTP server: %w", err)
				}
				return ftpCon, nil
			},
			Close: func(i interface{}) error {
				return i.(*ftp.ServerConn).Quit()
			},
			Ping:        nil,
			IdleTimeout: 0,
		}
		var p pool.Pool
		p, err = pool.NewChannelPool(f.ftpConfig)
		if err != nil {
			err = fmt.Errorf("Connection::Connect error while creating new channel pool: %w", err)
			return
		}
		f.ftpPool = p
	})
	return
}

func (f *Connection) Close() {
	if f.ftpPool == nil {
		return
	}
	f.ftpPool.Release()
}

func (f *Connection) FileSize(_ context.Context, path string) (int64, error) {
	con, err := f.ftpPool.Get()
	if err != nil {
		return 0, fmt.Errorf("Connection::FileSize error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	res, err := ftpCon.FileSize(path)
	if err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return 0, ErrorFtpPermissionDenied
		}
		return 0, fmt.Errorf("Connection::FileSize error while getting file size: %w", err)
	}
	return res, nil
}

func (f *Connection) Delete(_ context.Context, path string) error {
	con, err := f.ftpPool.Get()
	if err != nil {
		return fmt.Errorf("Connection::Delete error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	if err := ftpCon.Delete(path); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return ErrorFtpPermissionDenied
		}
		return fmt.Errorf("Connection::Delete error while deleting file %s: %w", path, err)
	}
	return nil
}

func (f *Connection) RemoveDirRecur(_ context.Context, path string) error {
	con, err := f.ftpPool.Get()
	if err != nil {
		return fmt.Errorf("Connection::RemoveDirRecur error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	if err := ftpCon.RemoveDirRecur(path); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return ErrorFtpPermissionDenied
		}
		return fmt.Errorf("Connection::RemoveDirRecur error while deleting directory %s: %w", path, err)
	}
	return nil
}

func (f *Connection) List(_ context.Context, dir string) ([]*ftp.Entry, error) {
	con, err := f.ftpPool.Get()
	if err != nil {
		return nil, fmt.Errorf("Connection::List error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	list, err := ftpCon.List(dir)
	if err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return nil, ErrorFtpPermissionDenied
		}
		return nil, fmt.Errorf("Connection::List error while listing directory %s: %w", dir, err)
	}
	return list, nil
}

func (f *Connection) Walk(_ context.Context, dir string) ([]*ftp.Entry, error) {
	con, err := f.ftpPool.Get()
	if err != nil {
		return nil, fmt.Errorf("Connection::Walk error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	res := make([]*ftp.Entry, 0, 500)
	walker := ftpCon.Walk(dir)
	for walker.Next() {
		res = append(res, walker.Stat())
	}
	if err := walker.Err(); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return nil, ErrorFtpPermissionDenied
		}
		return nil, fmt.Errorf("Connection::Walk error while listing directory %s: %w", dir, err)
	}
	return res, nil
}

func (f *Connection) Retr(_ context.Context, path string) (*ftp.Response, error) {
	con, err := f.ftpPool.Get()
	if err != nil {
		return nil, fmt.Errorf("FtpReader::Retr error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	response, err := ftpCon.Retr(path)
	if err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return nil, ErrorFtpPermissionDenied
		}
		return nil, fmt.Errorf("FtpReader::Retr error while reading file %s: %w", path, err)
	}
	return response, nil
}

func (f *Connection) Stor(_ context.Context, path string, r io.Reader) error {
	con, err := f.ftpPool.Get()
	if err != nil {
		return fmt.Errorf("FtpReader::Stor error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	if err := ftpCon.Stor(path, r); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return ErrorFtpPermissionDenied
		}
		return fmt.Errorf("Connection::Stor error while uploading file %s: %w", path, err)
	}
	return nil
}

func (f *Connection) MakeDir(_ context.Context, path string) error {
	con, err := f.ftpPool.Get()
	if err != nil {
		return fmt.Errorf("FtpReader::MakeDir error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	if err := ftpCon.MakeDir(path); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return ErrorFtpPermissionDenied
		}
		return fmt.Errorf("Connection::MakeDir error while creating directory %s: %w", path, err)
	}
	return nil
}

func (f *Connection) Chmod(_ context.Context, path string, mode string) error {
	con, err := f.ftpPool.Get()
	if err != nil {
		return fmt.Errorf("FtpReader::Chmod error while getting connection from pool: %w", err)
	}
	defer func() {
		_ = f.ftpPool.Put(con)
	}()
	ftpCon := con.(*ftp.ServerConn)
	if err := ftpCon.Chmod(path, mode); err != nil {
		if err.(*textproto.Error).Code == ftpPermissionDeniedCode {
			return ErrorFtpPermissionDenied
		}
		return fmt.Errorf("Connection::Chmod error while changing permissions for %s: %w", path, err)
	}
	return nil
}
