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

	if err := convertError(result.Unexpected); err != nil {
		return err
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

func convertError(unexpected *domain.UnexpectedStruct) error {
	if unexpected == nil {
		return nil
	}
	messages := make([]string, 0)
	if unexpected.DuplicatedMetadata != nil && len(unexpected.DuplicatedMetadata) > 0 {
		m := "Duplicated metadata are found.\n"
		for _, path := range unexpected.DuplicatedMetadata {
			m += fmt.Sprintf("- %s\n", path)
		}
		messages = append(messages, m)
	}
	if unexpected.DuplicatedMedia != nil && len(unexpected.DuplicatedMedia) > 0 {
		m := "Duplicated medias are found.\n"
		for _, path := range unexpected.DuplicatedMedia {
			m += fmt.Sprintf("- %s\n", path)
		}
		messages = append(messages, m)
	}
	if unexpected.NotFoundMetadata != nil && len(unexpected.NotFoundMetadata) > 0 {
		m := "Not found metadata.\n"
		for _, path := range unexpected.NotFoundMetadata {
			m += fmt.Sprintf("- %s\n", path)
		}
		messages = append(messages, m)
	}
	if unexpected.NotFoundMedia != nil && len(unexpected.NotFoundMedia) > 0 {
		m := "Not found media file.\n"
		for _, path := range unexpected.NotFoundMedia {
			m += fmt.Sprintf("- %s\n", path)
		}
		messages = append(messages, m)
	}
	return fmt.Errorf("Unexpected analysis result\n%s", strings.Join(messages, "\n\n"))
}
