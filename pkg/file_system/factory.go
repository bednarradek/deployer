package file_system

import "github.com/bednarradek/php-deployer/pkg/ftp"

type FtpFactory struct {
	connection        *ftp.Connection
	defaultFileMode   string
	defaultFolderMode string
}

func NewFtpFactory(
	connection *ftp.Connection,
	defaultFileMode string,
	defaultFolderMode string,
) *FtpFactory {
	return &FtpFactory{
		connection:        connection,
		defaultFileMode:   defaultFileMode,
		defaultFolderMode: defaultFolderMode,
	}
}

func (f *FtpFactory) Creator() Creator {
	return NewFtpCreator(f.connection, f.defaultFolderMode)
}

func (f *FtpFactory) Deleter() Deleter {
	return NewFtpDeleter(f.connection)
}

func (f *FtpFactory) Lister() Lister {
	return NewFtpLister(f.connection)
}

func (f *FtpFactory) RecursiveLister() Lister {
	return NewRecursiveLister(f.Lister())
}

func (f *FtpFactory) Reader() Reader {
	return NewFtpReader(f.connection)
}

func (f *FtpFactory) HashReader() HashReader {
	return NewStandardHashReader(f.Reader())
}

func (f *FtpFactory) CompressionReader() Reader {
	return NewCompressionReader(f.Reader())
}

func (f *FtpFactory) Writer() Writer {
	return NewFtpWriter(f.connection, f.defaultFileMode)
}

func (f *FtpFactory) CompressionWriter() Writer {
	return NewCompressionWriter(f.Writer())
}

func (f *FtpFactory) ChangeModer() ChangeModer {
	return NewFtpChangeModer(f.connection)
}

type SystemFactory struct {
	defaultFileMode   string
	defaultFolderMode string
}

func NewSystemFactory(
	defaultFileMode string,
	defaultFolderMode string,
) *SystemFactory {
	return &SystemFactory{
		defaultFileMode:   defaultFileMode,
		defaultFolderMode: defaultFolderMode,
	}
}

func (s *SystemFactory) Deleter() Deleter {
	return NewSystemDeleter()
}

func (s *SystemFactory) Lister() Lister {
	return NewSystemLister()
}

func (s *SystemFactory) RecursiveLister() Lister {
	return NewRecursiveLister(s.Lister())
}

func (s *SystemFactory) Reader() Reader {
	return NewSystemReader()
}

func (s *SystemFactory) HashReader() HashReader {
	return NewStandardHashReader(s.Reader())
}

func (s *SystemFactory) Writer() Writer {
	return NewSystemWriter()
}

type LogFactory struct {
	logPath string
}

func NewLogFactory(logPath string) *LogFactory {
	return &LogFactory{logPath: logPath}
}

func (l *LogFactory) Lister(lister Lister, reader Reader) *LogLister {
	return NewLogLister(l.logPath, lister, reader)
}

func (l *LogFactory) HashReader(logFile LogFile, reader HashReader) HashReader {
	return NewLogHashReader(logFile, reader)
}
