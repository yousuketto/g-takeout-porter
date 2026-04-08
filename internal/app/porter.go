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
		messages := make([]string, 0)
		if result.Unexpected.DuplicatedMetadata != nil && len(result.Unexpected.DuplicatedMetadata) > 0 {
			m := "Duplicated metadata are found.\n"
			for _, path := range result.Unexpected.DuplicatedMetadata {
				m += fmt.Sprintf("- %s\n", path)
			}
			messages = append(messages, m)
		}
		if result.Unexpected.DuplicatedMedia != nil && len(result.Unexpected.DuplicatedMedia) > 0 {
			m := "Duplicated medias are found.\n"
			for _, path := range result.Unexpected.DuplicatedMedia {
				m += fmt.Sprintf("- %s\n", path)
			}
			messages = append(messages, m)
		}
		if result.Unexpected.NotFoundMetadata != nil && len(result.Unexpected.NotFoundMetadata) > 0 {
			m := "Not found metadata.\n"
			for _, path := range result.Unexpected.NotFoundMetadata {
				m += fmt.Sprintf("- %s\n", path)
			}
			messages = append(messages, m)
		}
		if result.Unexpected.NotFoundMedia != nil && len(result.Unexpected.NotFoundMedia) > 0 {
			m := "Not found media file.\n"
			for _, path := range result.Unexpected.NotFoundMedia {
				m += fmt.Sprintf("- %s\n", path)
			}
			messages = append(messages, m)
		}
		return fmt.Errorf("Unexpected analysis result\n%s", strings.Join(messages, "\n\n"))
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
