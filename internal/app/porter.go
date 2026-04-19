package app

import (
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/domain"
	"strings"
)

type porter struct {
	mediaMetadataRepo domain.MediaMetadataRepo
	backupStorage     domain.BackupStorage
}

func NewPorter(mediaMetadataRepo domain.MediaMetadataRepo, backupStorage domain.BackupStorage) *porter {
	return &porter{mediaMetadataRepo, backupStorage}
}

func (p *porter) Run(sourceDir, destDir string) error {
	result, err := p.mediaMetadataRepo.AnalyzeAllMetadata(sourceDir)
	if err != nil {
		return err
	}

	if err := convertError(result.Unexpected); err != nil {
		return err
	}

	copiedResults, err := p.backupStorage.Copy(result.Medias, destDir)
	if err != nil {
		return err
	}
	var failedReulsts []string
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

func (p *porter) DryRun(sourceDir, destDir string) ([]string, error) {
	result, err := p.mediaMetadataRepo.AnalyzeAllMetadata(sourceDir)
	if err != nil {
		return nil, err
	}

	if err := convertError(result.Unexpected); err != nil {
		return nil, err
	}

	results := p.backupStorage.DryCopy(result.Medias, destDir)
	pathInformation := make([]string, 0, len(results))
	for _, r := range results {
		pathInformation = append(pathInformation, fmt.Sprintf("%s -> %s", r.From, r.To))
	}
	return pathInformation, err
}

func convertError(unexpected *domain.UnexpectedStruct) error {
	if unexpected == nil {
		return nil
	}
	var messages []string
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
