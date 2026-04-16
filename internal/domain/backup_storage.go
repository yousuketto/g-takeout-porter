package domain

type CopiedResult struct {
	IsSuccess bool
	Media     MediaMetadata
}

type DryCopiedResult struct {
	From string
	To   string
}

type BackupStorage interface {
	Copy(sourceMetadata []MediaMetadata, destDir string) ([]CopiedResult, error)
	DryCopy(sourceMetadata []MediaMetadata, destDir string) []DryCopiedResult
}
