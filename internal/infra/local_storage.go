package infra

import (
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/domain"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LocalStorage struct {
	numWorkers int
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{numWorkers: 8}
}

func (storage *LocalStorage) Copy(sourceMetadata []domain.MediaMetadata, destDir string) ([]domain.CopiedResult, error) {
	totalLen := len(sourceMetadata)
	destPaths := make([]string, 0, totalLen)
	for _, metadata := range sourceMetadata {
		path := destPath(destDir, metadata)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		destPaths = append(destPaths, path)
	}

	results := make([]domain.CopiedResult, totalLen)

	jobs := make(chan int, totalLen)
	for i := 0; i < totalLen; i++ {
		jobs <- i
	}
	close(jobs)

	numWorkers := storage.numWorkers
	if totalLen < numWorkers {
		numWorkers = totalLen
	}
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				metadata := sourceMetadata[i]
				path := destPaths[i]
				err := copyFile(metadata.RelativePath, path)
				if err != nil {
					fmt.Printf("Fail to copy '%s' to '%s': %v\n", metadata.RelativePath, path, err)
					results[i] = domain.CopiedResult{IsSuccess: false, Media: metadata}
					continue
				}

				t := time.Unix(metadata.Timestamp, 0)
				if err := os.Chtimes(path, t, t); err != nil {
					fmt.Printf("Fail to set time for '%s': %v\n", path, err)
					results[i] = domain.CopiedResult{IsSuccess: false, Media: metadata}
					continue
				}
				results[i] = domain.CopiedResult{IsSuccess: true, Media: metadata}
			}
		}()
	}
	wg.Wait()
	return results, nil
}

func (storage *LocalStorage) DryCopy(sourceMetadata []domain.MediaMetadata, destDir string) []domain.DryCopiedResult {
	results := make([]domain.DryCopiedResult, 0, len(sourceMetadata))
	for _, metadata := range sourceMetadata {
		results = append(results, domain.DryCopiedResult{metadata.RelativePath, destPath(destDir, metadata)})
	}
	return results
}

func destPath(destDir string, metadata domain.MediaMetadata) string {
	timestamp := time.Unix(metadata.Timestamp, 0)
	return filepath.Join(destDir, timestamp.Format("200601"), filepath.Base(metadata.RelativePath))
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}
