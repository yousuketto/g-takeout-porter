package app

import (
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/domain"
	"strings"
)

type Porter struct {
	mediaMetadataRepo domain.MediaMetadataRepo
	backupStorage     domain.BackupStorage
}

func NewPorter(mediaMetadataRepo domain.MediaMetadataRepo, backupStorage domain.BackupStorage) *Porter {
	return &Porter{mediaMetadataRepo, backupStorage}
}

func (porter *Porter) Run(sourceDir, destDir string) error {
	result, err := porter.mediaMetadataRepo.AnalyzeAllMetadata(sourceDir)
	if err != nil {
		return err
	}

	if result.Unexpected != nil {
		return fmt.Errorf("Unexpected analysis result: %v", result.Unexpected)
	}

	copiedResults, err := porter.backupStorage.Copy(result.Medias, destDir)
	if err != nil {
		return err
	}
	failedReulsts := make([]string, 0)
	for _, result := range copiedResults {
		if !result.IsSuccess {
			failedReulsts = append(failedReulsts, result.Media.RelativePath)
		}
	}
	if len(failedReulsts) > 0 {
		return fmt.Errorf("Fail to copy (%s)", strings.Join(failedReulsts, ", "))
	}

	return nil
}
