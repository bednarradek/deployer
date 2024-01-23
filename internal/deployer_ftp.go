package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bednarradek/php-deployer/pkg/action"
	"github.com/bednarradek/php-deployer/pkg/file_system"
	"github.com/bednarradek/php-deployer/pkg/filter"
	"github.com/bednarradek/php-deployer/pkg/ftp"
	"github.com/bednarradek/php-deployer/pkg/generator"
	"github.com/bednarradek/php-deployer/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type FtpDeployer struct {
	config           *FtpConfig
	ftpConnection    *ftp.Connection
	ftpFactory       *file_system.FtpFactory
	systemFactory    *file_system.SystemFactory
	logFactory       *file_system.LogFactory
	fileSystemFilter filter.Filter
	envGenerator     generator.Generator
}

func NewFtpDeployer(config *FtpConfig) (*FtpDeployer, error) {
	// connect to ftp
	logrus.Infof("Connecting to FTP server...")

	//create env generator
	envGenerator := generator.NewEnvironmentGenerator()
	user, err := envGenerator.Generate(context.Background(), []byte(config.Sync.FtpConfig.User))
	if err != nil {
		return nil, fmt.Errorf("FtpDeployer::NewFtpDeployer error while generating user: %w", err)
	}

	password, err := envGenerator.Generate(context.Background(), []byte(config.Sync.FtpConfig.Password))
	if err != nil {
		return nil, fmt.Errorf("FtpDeployer::NewFtpDeployer error while generating password: %w", err)
	}

	ftpConnection := ftp.NewConnection(
		config.Sync.FtpConfig.Host,
		string(user),
		string(password),
	)
	if err := ftpConnection.Connect(); err != nil {
		return nil, fmt.Errorf("FtpDeployer::Deploy error while connecting to ftp: %w", err)
	}

	// create FTP factory
	logrus.Infof("Preparing all necessary objects...")

	ftpFactory := file_system.NewFtpFactory(
		ftpConnection,
		config.Sync.DefaultFileMode,
		config.Sync.DefaultDirMode,
	)

	// create System factory - default mode does not matter in this case
	systemFactory := file_system.NewSystemFactory(
		"775",
		"0775",
	)

	// create Log factory
	logFactory := file_system.NewLogFactory(config.Sync.LogFileDest)

	// create filter
	fileSystemFilter := filter.NewFileSystemFilter(config.Sync.IgnoreList)

	return &FtpDeployer{
		config:           config,
		ftpConnection:    ftpConnection,
		ftpFactory:       ftpFactory,
		systemFactory:    systemFactory,
		logFactory:       logFactory,
		fileSystemFilter: fileSystemFilter,
		envGenerator:     envGenerator,
	}, nil
}

func (f *FtpDeployer) Close() {
	f.ftpConnection.Close()
}

// FtpDeployer::doGenerate [EnvironmentGenerator] generate file from template and save it to destination on local
func (f *FtpDeployer) doGenerate(ctx context.Context, generatorConfig GeneratorConfig) error {
	switch generatorConfig.Type {
	case EnvironmentGenerator:
		templatePath, ok := generatorConfig.Arguments["templatePath"]
		if !ok {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] missing templatePath argument for generator: %s", generatorConfig.Type)
		}
		destination, ok := generatorConfig.Arguments["destination"]
		if !ok {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] missing destination argument for generator: %s", generatorConfig.Type)
		}
		tmpl, err := f.systemFactory.Reader().Read(ctx, fmt.Sprintf("%s%s", f.config.Sync.Source, templatePath))
		if err != nil {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] error while reading template file %s: %w", templatePath, err)
		}
		res, err := f.envGenerator.Generate(ctx, tmpl)
		if err != nil {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] error while generating environment file: %w", err)
		}
		destPath := fmt.Sprintf("%s%s", f.config.Sync.Source, destination)
		if err := f.systemFactory.Deleter().Delete(ctx, destPath); err != nil {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] error while deleting environment file %s: %w", destPath, err)
		}
		if err := f.systemFactory.Writer().Write(ctx, destPath, res); err != nil {
			return fmt.Errorf("FtpDeployer::doGenerate [EnvironmentGenerator] error while writing environment file %s: %w", destination, err)
		}
		return nil
	default:
		logrus.Warningf("Unknown generator type: %s", generatorConfig.Type)
	}
	return nil
}

// FtpDeployer::doMove move file from source to destination on ftp
func (f *FtpDeployer) doMove(ctx context.Context, moverConfig MoveConfig) error {
	content, err := f.systemFactory.Reader().Read(ctx, fmt.Sprintf("%s%s", f.config.Sync.Source, moverConfig.Source))
	if err != nil {
		return fmt.Errorf("FtpDeployer::doMove error while reading file %s: %w", fmt.Sprintf("%s%s", f.config.Sync.Source, moverConfig.Source), err)
	}
	dest := fmt.Sprintf("%s%s", f.config.Sync.Destination, moverConfig.Destination)
	if err := f.ftpFactory.Creator().CreateDir(ctx, helpers.GetDirectoryPath(dest)); err != nil {
		return fmt.Errorf("FtpDeployer::doMove error while creating directory %s: %w", helpers.GetDirectoryPath(dest), err)
	}
	if err := f.ftpFactory.Writer().Write(ctx, dest, content); err != nil {
		return fmt.Errorf("FtpDeployer::doMove error while writing file %s: %w", dest, err)
	}
	return nil
}

// FtpDeployer::doAction [HttpAction] call http request
func (f *FtpDeployer) doAction(ctx context.Context, actionConfig ActionConfig) error {
	switch actionConfig.Type {
	case HttpAction:
		urlArg, ok := actionConfig.Arguments["url"]
		if !ok {
			return fmt.Errorf("FtpDeployer::doAction [HttpAction] missing url argument for action: %s", actionConfig.Type)
		}
		url, err := f.envGenerator.Generate(ctx, []byte(fmt.Sprintf("%s", urlArg)))
		if err != nil {
			return fmt.Errorf("FtpDeployer::doAction [HttpAction] error while generating url: %w", err)
		}

		methodArg, ok := actionConfig.Arguments["method"]
		if !ok {
			return fmt.Errorf("FtpDeployer::doAction [HttpAction] missing method argument for action: %s", actionConfig.Type)
		}
		method, err := f.envGenerator.Generate(ctx, []byte(fmt.Sprintf("%s", methodArg)))
		if err != nil {
			return fmt.Errorf("FtpDeployer::doAction [HttpAction] error while generating method: %w", err)
		}

		headers := make(map[string]string)
		headersArg, ok := actionConfig.Arguments["headers"]
		if ok {
			for k, v := range headersArg.(map[string]interface{}) {
				res, err := f.envGenerator.Generate(ctx, []byte(fmt.Sprintf("%s", v)))
				if err != nil {
					return fmt.Errorf("FtpDeployer::doAction [HttpAction] error while generating header %s: %w", k, err)
				}
				headers[k] = string(res)
			}
		}

		body := new(bytes.Buffer)
		bodyArg, ok := actionConfig.Arguments["body"]
		if ok {
			res, err := f.envGenerator.Generate(ctx, []byte(fmt.Sprintf("%s", bodyArg)))
			if err != nil {
				return fmt.Errorf("FtpDeployer::doAction [HttpAction] error while generating body: %w", err)
			}
			body.Write(res)
		}

		if err := action.NewHttpAction(string(url), string(method), headers, body).Do(ctx); err != nil {
			return fmt.Errorf("FtpDeployer::doAction [HttpAction] error while calling http request: %w", err)
		}
		return nil
	default:
		logrus.Warningf("Unknown action type: %s", actionConfig.Type)
	}
	return nil
}

// FtpDeployer::doClean delete files and folders on FTP
func (f *FtpDeployer) doClean(ctx context.Context, cleanConfig CleanConfig) error {
	for _, file := range cleanConfig.Files {
		if err := f.ftpFactory.Deleter().Delete(ctx, fmt.Sprintf("%s%s", f.config.Sync.Destination, file)); err != nil {
			return fmt.Errorf("FtpDeployer::doClean error while deleting file %s: %w", file, err)
		}
	}
	for _, folder := range cleanConfig.Folders {
		content, err := f.ftpFactory.Lister().List(ctx, fmt.Sprintf("%s%s", f.config.Sync.Destination, folder))
		if err != nil {
			return fmt.Errorf("FtpDeployer::doClean error while listing folder %s: %w", folder, err)
		}
		for _, c := range content {
			p := fmt.Sprintf("%s%s/%s", f.config.Sync.Destination, folder, c.GetName())
			if c.IsDir() {
				if err := f.ftpFactory.Deleter().DeleteDir(ctx, p); err != nil {
					return fmt.Errorf("FtpDeployer::doClean error while deleting folder %s: %w", folder, err)
				}
				continue
			}
			if err := f.ftpFactory.Deleter().Delete(ctx, p); err != nil {
				return fmt.Errorf("FtpDeployer::doClean error while deleting file %s: %w", folder, err)
			}
		}

	}
	return nil
}

func (f *FtpDeployer) doStep(ctx context.Context, stepConfig StepConfig) error {
	// generator
	for _, c := range stepConfig.Generate {
		if err := f.doGenerate(ctx, c); err != nil {
			return err
		}
	}

	// move
	for _, c := range stepConfig.Move {
		if err := f.doMove(ctx, c); err != nil {
			return err
		}
	}

	// action
	for _, c := range stepConfig.Action {
		if err := f.doAction(ctx, c); err != nil {
			return err
		}
	}

	// deleter
	if err := f.doClean(ctx, stepConfig.Clean); err != nil {
		return err
	}

	return nil
}

func (f *FtpDeployer) sync(ctx context.Context) error {
	// read system objects
	systemObjects, err := NewReaderManager(
		f.systemFactory.RecursiveLister(),
		f.systemFactory.HashReader(),
		f.fileSystemFilter,
	).Read(ctx, f.config.Sync.Source)

	if err != nil {
		return fmt.Errorf("FtpDeployer::sync error while reading system objects: %w", err)
	}

	// read ftp objects
	logLister := f.logFactory.Lister(
		f.ftpFactory.RecursiveLister(),
		f.ftpFactory.CompressionReader(),
	)
	logFile, err := logLister.GetLogFile(ctx)
	if err != nil {
		return fmt.Errorf("FtpDeployer::sync error while reading log file: %w", err)
	}
	var finalFtpHashReader file_system.HashReader
	if logFile != nil {
		finalFtpHashReader = f.logFactory.HashReader(*logFile, f.ftpFactory.HashReader())
	} else {
		finalFtpHashReader = f.ftpFactory.HashReader()
	}

	ftpObjects, err := NewReaderManager(
		logLister,
		finalFtpHashReader,
		f.fileSystemFilter,
	).Read(ctx, f.config.Sync.Destination)

	if err != nil {
		return fmt.Errorf("FtpDeployer::sync error while reading ftp objects: %w", err)
	}

	// convert object to map
	mapSystemObjects := helpers.ConvertToMap(systemObjects)
	mapFtpObjects := helpers.ConvertToMap(ftpObjects)

	// compare objects
	diff := NewCompareManager().Compare(mapSystemObjects, mapFtpObjects)

	// resolve diff files
	if err := NewResolverManager(
		f.systemFactory.Reader(),
		f.ftpFactory.Writer(),
		f.ftpFactory.Creator(),
		f.ftpFactory.Deleter(),
		f.config.Sync.Source,
		f.config.Sync.Destination,
	).Resolve(ctx, diff); err != nil {
		return fmt.Errorf("FtpDeployer::sync error while resolving: %w", err)
	}

	// upload log file
	logObjects := make([]file_system.LogObject, len(systemObjects))
	for i, o := range systemObjects {
		logObjects[i] = file_system.LogObject{
			Path:          o.Path(),
			IsDirFlag:     o.IsDir(),
			IsRegularFlag: !o.IsDir(),
			Hash:          o.Hash(),
		}
	}

	// marshal objects
	b, err := json.Marshal(file_system.LogFile{Objects: logObjects})
	if err != nil {
		return fmt.Errorf("FtpDeployer::sync error while marshalling log file: %w", err)
	}

	// create log file directory
	if err := f.ftpFactory.Creator().CreateDir(ctx, helpers.GetDirectoryPath(f.config.Sync.LogFileDest)); err != nil {
		return fmt.Errorf("FtpDeployer::sync error while creating log file directory: %w", err)
	}

	if err := f.ftpFactory.
		CompressionWriter().
		Write(
			ctx,
			f.config.Sync.LogFileDest,
			b,
		); err != nil {
		return fmt.Errorf("FtpDeployer::sync error while writing log file: %w", err)
	}

	return nil
}

func (f *FtpDeployer) folders(ctx context.Context, folders []string) error {
	for _, fol := range folders {
		if err := f.ftpFactory.Creator().CreateDir(ctx, fol); err != nil {
			return fmt.Errorf("FtpDeployer::folders error while creating folder %s: %w", fol, err)
		}
	}
	return nil
}

func (f *FtpDeployer) readableFolders(ctx context.Context, folders []string) error {
	for _, fol := range folders {
		res, err := f.ftpFactory.RecursiveLister().List(ctx, fmt.Sprintf("%s%s", f.config.Sync.Destination, fol))
		if err != nil {
			return fmt.Errorf("FtpDeployer::readableFolders error while listing folder %s: %w", fol, err)
		}
		for _, r := range res {
			abs := fmt.Sprintf("%s%s", f.config.Sync.Destination, r.GetName())
			if r.IsDir() {
				if err := f.ftpFactory.ChangeModer().Change(ctx, abs, "0777"); err != nil {
					return fmt.Errorf("FtpDeployer::readableFolders error while changing mode of folder %s: %w", fol, err)
				}
				continue
			}
			if err := f.ftpFactory.ChangeModer().Change(ctx, abs, "0777"); err != nil {
				return fmt.Errorf("FtpDeployer::readableFolders error while changing mode of file %s: %w", fol, err)
			}
		}
	}
	return nil
}

func (f *FtpDeployer) Deploy(ctx context.Context) error {
	// prerequisites:
	// - installed composer
	// - installed node modules
	// - build assets

	// steps:
	// - setup maitinance mode DONE - load and copy
	// - upload files DONE
	// - delete cache DONE - deleter
	// - run migrations DONe - caller
	// - remove maintenance mode - deleter
	// - what about dynamic created files in www folder?

	//-------- start of final solution

	// do step before
	if err := f.doStep(ctx, f.config.Before); err != nil {
		return fmt.Errorf("FtpDeployer::Deploy error while executing before step: %w", err)
	}

	// do sync
	if err := f.sync(ctx); err != nil {
		return fmt.Errorf("FtpDeployer::Deploy error while syncing: %w", err)
	}

	// do folders
	if err := f.folders(ctx, f.config.Folders); err != nil {
		return fmt.Errorf("FtpDeployer::Deploy error while creating folders: %w", err)
	}

	// do readable folders
	if err := f.readableFolders(ctx, f.config.ReadableFolders); err != nil {
		return fmt.Errorf("FtpDeployer::Deploy error while changing mode of folders: %w", err)
	}

	// do step after
	if err := f.doStep(ctx, f.config.After); err != nil {
		return fmt.Errorf("FtpDeployer::Deploy error while executing after step: %w", err)
	}

	return nil
}
