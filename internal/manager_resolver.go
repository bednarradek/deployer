package internal

import (
	"context"
	"fmt"

	"github.com/bednarradek/php-deployer/pkg/file_system"
	"github.com/bednarradek/php-deployer/pkg/helpers"
)

type ResolverManager struct {
	localReader    file_system.Reader
	remoteUploader file_system.Writer
	remoteCreator  file_system.Creator
	remoteDeleter  file_system.Deleter
	localPath      string
	remotePath     string
}

func NewResolverManager(
	localReader file_system.Reader,
	remoteUploader file_system.Writer,
	remoteCreator file_system.Creator,
	remoteDeleter file_system.Deleter,
	localPath string,
	remotePath string,
) *ResolverManager {
	return &ResolverManager{
		localReader:    localReader,
		remoteUploader: remoteUploader,
		remoteCreator:  remoteCreator,
		remoteDeleter:  remoteDeleter,
		localPath:      localPath,
		remotePath:     remotePath,
	}
}

func (p *ResolverManager) Resolve(ctx context.Context, input []CompareResult) error {
	folders := make([]CompareResult, 0, 100)
	files := make([]CompareResult, 0, 100)
	for _, result := range input {
		if result.Object.IsDir() {
			folders = append(folders, result)
			continue
		}
		files = append(files, result)
	}
	if err := p.resolve(ctx, folders); err != nil {
		return fmt.Errorf("ResolverManager::Resolve error while resolving folders: %w", err)
	}
	if err := p.resolve(ctx, files); err != nil {
		return fmt.Errorf("ResolverManager::Resolve error while resolving files: %w", err)
	}
	return nil
}

func (p *ResolverManager) resolve(ctx context.Context, input []CompareResult) error {
	_, err := helpers.RunWorkers(ctx, 10, input, func(ctx context.Context, i CompareResult) (interface{}, error) {
		switch i.Action {
		case ActionChange:
			if err := p.delete(ctx, i.Object); err != nil {
				return nil, fmt.Errorf("ResolverManager::resolve error while deleting %s: %w", i.Object.Path(), err)
			}
			if err := p.upload(ctx, i.Object); err != nil {
				return nil, fmt.Errorf("ResolverManager::resolve error while uploading %s: %w", i.Object.Path(), err)
			}
		case ActionUpload:
			if err := p.upload(ctx, i.Object); err != nil {
				return nil, fmt.Errorf("ResolverManager::resolve error while uploading %s: %w", i.Object.Path(), err)
			}
		case ActionDelete:
			if err := p.delete(ctx, i.Object); err != nil {
				return nil, fmt.Errorf("ResolverManager::resolve error while deleting %s: %w", i.Object.Path(), err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("ResolverManager::resolve error while resolving: %w", err)
	}
	return nil
}

//func (p *ResolverManager) resolve(ctx context.Context, input []CompareResult) error {
//	numWorkers := 5
//	if len(input) < numWorkers {
//		numWorkers = len(input)
//	}
//	jobs := make(chan CompareResult, len(input))
//	errChan := make(chan error)
//	defer func() {
//		close(errChan)
//	}()
//	doneChan := make(chan struct{}, 1)
//	wg := sync.WaitGroup{}
//
//	for i := 0; i < numWorkers; i++ {
//		go func(ctx context.Context, job <-chan CompareResult, errChan chan<- error) {
//			wg.Add(1)
//			defer func() {
//				wg.Done()
//			}()
//			for result := range job {
//				switch result.Action {
//				case ActionChange:
//					if err := p.delete(ctx, result.Object); err != nil {
//						errChan <- fmt.Errorf("ResolverManager::resolve error while deleting %s: %w", result.Object.Path(), err)
//					}
//					if err := p.upload(ctx, result.Object); err != nil {
//						errChan <- fmt.Errorf("ResolverManager::resolve error while uploading %s: %w", result.Object.Path(), err)
//					}
//					continue
//				case ActionUpload:
//					if err := p.upload(ctx, result.Object); err != nil {
//						errChan <- fmt.Errorf("ResolverManager::resolve error while uploading %s: %w", result.Object.Path(), err)
//					}
//					continue
//				case ActionDelete:
//					if err := p.delete(ctx, result.Object); err != nil {
//						errChan <- fmt.Errorf("ResolverManager::resolve error while deleting %s: %w", result.Object.Path(), err)
//					}
//				}
//			}
//		}(ctx, jobs, errChan)
//	}
//
//	// send all files to workers
//	for _, file := range input {
//		jobs <- file
//	}
//	close(jobs)
//
//	// block until all workers are done
//	go func() {
//		wg.Wait()
//		close(doneChan)
//	}()
//
//	for {
//		select {
//		case err := <-errChan:
//			close(jobs)
//			close(doneChan)
//			return err
//		case <-doneChan:
//			return nil
//		}
//	}
//}

func (p *ResolverManager) upload(ctx context.Context, object CompareObject) error {
	if object.IsDir() {
		if err := p.remoteCreator.CreateDir(ctx, fmt.Sprintf("%s/%s", p.remotePath, object.Path())); err != nil {
			return fmt.Errorf("ResolverManager::upload error while creating directory %s: %w", object.Path(), err)
		}
		return nil
	}
	content, err := p.localReader.Read(ctx, fmt.Sprintf("%s/%s", p.localPath, object.Path()))
	if err != nil {
		return fmt.Errorf("ResolverManager::upload error while reading file %s: %w", object.Path(), err)
	}
	if err := p.remoteUploader.Write(ctx, fmt.Sprintf("%s/%s", p.remotePath, object.Path()), content); err != nil {
		return fmt.Errorf("ResolverManager::upload error while uploading file %s: %w", object.Path(), err)
	}
	return nil
}

func (p *ResolverManager) delete(ctx context.Context, object CompareObject) error {
	if object.IsDir() {
		if err := p.remoteDeleter.DeleteDir(ctx, fmt.Sprintf("%s/%s", p.remotePath, object.Path())); err != nil {
			return fmt.Errorf("ResolverManager::delete error while deleting directory %s: %w", object.Path(), err)
		}
		return nil
	}
	if err := p.remoteDeleter.Delete(ctx, fmt.Sprintf("%s/%s", p.remotePath, object.Path())); err != nil {
		return fmt.Errorf("ResolverManager::delete error while deleting file %s: %w", object.Path(), err)
	}
	return nil
}
