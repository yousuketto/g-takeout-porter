package infra

import (
	"encoding/json"
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/domain"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type TakeoutMetadataRepo struct{}

func NewTakeoutMetadataRepo() *TakeoutMetadataRepo {
	return &TakeoutMetadataRepo{}
}

type metadataJson struct {
	Title          string `json:"title"`
	PhotoTakenTime struct {
		Timestamp string `json:"timestamp"`
	} `json:"photoTakenTime"`
}

func (repo *TakeoutMetadataRepo) AnalyzeAllMetadata(dirPath string) (*domain.AnalysisResult, error) {
	timestampMap := make(map[string]int64)
	relativePathMap := make(map[string]string)
	var duplicatedMetadata []string
	var duplicatedFiles []string
	var notFoundMetadataPaths []string

	fallbackTimestampMap := make(map[string]int64)

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var m metadataJson
			if err := json.Unmarshal(data, &m); err != nil {
				return fmt.Errorf("Unexpected JSON (%s): %w", path, err)
			}
			if m.Title == "" {
				return fmt.Errorf("Not found tile in JSON (%s)", path)
			}
			if m.PhotoTakenTime.Timestamp == "" {
				return fmt.Errorf("Not found Timestamp in JSON (%s)", path)
			}

			takeoutRelativePath, err := extractTakeoutRelativePath(path)
			if err != nil {
				return err
			}

			ts, err := strconv.ParseInt(m.PhotoTakenTime.Timestamp, 10, 64)
			if err != nil {
				return fmt.Errorf("Fail to conver %s timestamp value: %w", m.Title, err)
			}

			key := filepath.Join(append(takeoutRelativePath[:len(takeoutRelativePath)-1], m.Title)...)
			if _, existing := timestampMap[key]; existing {
				duplicatedMetadata = append(duplicatedMetadata, key)
			} else {
				timestampMap[key] = ts
				fallbackTimestampMap[removeExt(key)] = ts
			}
		} else {
			takeoutRelativePath, err := extractTakeoutRelativePath(path)
			if err != nil {
				return err
			}

			key := filepath.Join(takeoutRelativePath...)
			if _, existing := relativePathMap[key]; existing {
				duplicatedFiles = append(duplicatedFiles, key)
			} else {
				relativePathMap[key] = path
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	mediaMetadataSlice := make([]domain.MediaMetadata, 0, len(relativePathMap))
	for k, relativePath := range relativePathMap {
		if timestamp, ok := timestampMap[k]; ok {
			// TODO Get RealTimestamp from exif
			mediaMetadataSlice = append(mediaMetadataSlice, domain.MediaMetadata{relativePath, timestamp, timestamp, false})
			delete(timestampMap, k)
		} else {
			if timestamp, ok := fallbackTimestampMap[removeExt(k)]; ok {
				// TODO Get RealTimestamp from exif
				mediaMetadataSlice = append(mediaMetadataSlice, domain.MediaMetadata{relativePath, timestamp, timestamp, true})
			} else {
				notFoundMetadataPaths = append(notFoundMetadataPaths, k)
			}
		}
	}

	if len(duplicatedMetadata) > 0 || len(duplicatedFiles) > 0 || len(timestampMap) > 0 || len(notFoundMetadataPaths) > 0 {
		return &domain.AnalysisResult{
			Medias: mediaMetadataSlice,
			Unexpected: &domain.UnexpectedStruct{
				duplicatedMetadata,
				duplicatedFiles,
				notFoundMetadataPaths,
				slices.Collect(maps.Keys(timestampMap)),
			},
		}, nil
	}
	return &domain.AnalysisResult{mediaMetadataSlice, nil}, nil
}

func extractTakeoutRelativePath(path string) ([]string, error) {
	parts := strings.Split(path, string(filepath.Separator))
	index := -1
	for i := len(parts) - 2; i >= 0; i-- {
		if parts[i] == "Takeout" {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("The path (%s) doesn't have Takeout.", path)
	}
	return parts[index:], nil
}

func removeExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}
