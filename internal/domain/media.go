package domain

type MediaMetadata struct {
	RelativePath          string
	Timestamp             int64
	RealTimestamp         int64
	IsFallbackedTimestamp bool
}

type UnexpectedStruct struct {
	DuplicatedMetadata []string
	DuplicatedMedia    []string
	NotFoundMetadata   []string
	NotFoundMedia      []string
}

type AnalysisResult struct {
	Medias     []MediaMetadata
	Unexpected *UnexpectedStruct
}

type MediaMetadataRepo interface {
	AnalyzeAllMetadata(dirPath string) (*AnalysisResult, error)
}
