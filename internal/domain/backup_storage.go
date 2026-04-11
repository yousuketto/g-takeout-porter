package domain

type CopiedResult struct {
	IsSuccess bool
	Media     MediaMetadata
}

type BackupStorage interface {
	Copy(sourceMetadata []MediaMetadata, destDir string) ([]CopiedResult, error)
}
