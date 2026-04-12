package infra

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yousuketto/g-takeout-porter/internal/domain"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestExtractTakeoutRelativePath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []string
		wantError bool
	}{
		{
			name:      "Positive: The path contains one 'Takeout' directory",
			input:     filepath.Join("some", "dir", "Takeout", "Google フォト", "photo.jpg"),
			want:      []string{"Takeout", "Google フォト", "photo.jpg"},
			wantError: false,
		},
		{
			name:      "Positive: The path contains two or more 'Takeout' directories",
			input:     filepath.Join("first", "Takeout", "second", "Takeout", "directory", "photo.img"),
			want:      []string{"Takeout", "directory", "photo.img"},
			wantError: false,
		},
		{
			name:      "Negative: The path doesn't contain a 'Takeout' directory",
			input:     filepath.Join("some", "dir", "Takeouts", "Google フォト", "photo.jpg"),
			want:      nil,
			wantError: true,
		},
		{
			name:      "Negative: The path doesn't contain a 'Takeout' directory, but contains a 'Takeout' file",
			input:     filepath.Join("some", "dir", "Takeouts", "Google フォト", "photo.jpg"),
			want:      nil,
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := extractTakeoutRelativePath(testCase.input)
			if (err != nil) != testCase.wantError {
				t.Errorf("error = %v, wantError %v", err, testCase.wantError)
			}
			if !slices.Equal(got, testCase.want) {
				t.Errorf("got = %v, want %v", got, testCase.want)
			}
		})

	}
}

func TestTakeoutMetadataRepo_AnalyzeAllMetadata_validStructure(t *testing.T) {
	temp := t.TempDir()
	files := map[string][]byte{
		// media and metadata in dir1
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.jpg"):  []byte(""),
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.json"): makeJson("photo1.jpg", "1397367114"),
		// media in dir1 and metadata in dir2
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo2.jpg"):  []byte(""),
		filepath.Join("dir2", "Takeout", "Google フォト", "2014", "photo2.json"): makeJson("photo2.jpg", "1397367111"),
		// media in dir1 and metadata is nothing
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "no_metadata.jpg"): []byte(""),
		// media is nothing and metadata in dir1
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "no_media.json"): makeJson("no_media.jpg", "1397367112"),
		// one media and two metadata
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo3.jpg"):  []byte(""),
		filepath.Join("dir2", "Takeout", "Google フォト", "2014", "photo3.json"): makeJson("photo3.jpg", "1397367113"),
		filepath.Join("dir3", "Takeout", "Google フォト", "2014", "photo3.json"): makeJson("photo3.jpg", "1397367113"),
		// two media and one metadata
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo4.jpg"):  []byte(""),
		filepath.Join("dir2", "Takeout", "Google フォト", "2014", "photo4.jpg"):  []byte(""),
		filepath.Join("dir3", "Takeout", "Google フォト", "2014", "photo4.json"): makeJson("photo4.jpg", "1397367115"),
		// one media file and one fallbacked file
		filepath.Join("dir2", "Takeout", "Google フォト", "2014", "photo5.jpg"):  []byte(""),
		filepath.Join("dir2", "Takeout", "Google フォト", "2014", "photo5"):      []byte(""),
		filepath.Join("dir3", "Takeout", "Google フォト", "2014", "photo5.json"): makeJson("photo5.jpg", "1397367116"),
	}
	for path, content := range files {
		fullPath := filepath.Join(temp, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	repo := NewTakeoutMetadataRepo()
	result, err := repo.AnalyzeAllMetadata(temp)

	if err != nil {
		t.Fatalf("AnalyzeAllMetadata returned error: %v", err)
	}
	if len(result.Medias) != 6 {
		t.Fatalf("AnalyzeAllMetadata returned %d medias, expected %d", len(result.Medias), 4)
		return
	}
	slices.SortFunc(result.Medias, func(a, b domain.MediaMetadata) int {
		if n := cmp.Compare(a.Timestamp, b.Timestamp); n != 0 {
			return n
		}
		return cmp.Compare(a.RelativePath, b.RelativePath)
	})
	// Check result.Medias
	wantPath := filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo2.jpg")
	if result.Medias[0].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[0].RelativePath)
	}
	if result.Medias[0].Timestamp != 1397367111 {
		t.Errorf("Expected timestamp %d, got %d", 1397367111, result.Medias[0].Timestamp)
	}
	if result.Medias[0].IsFallbackedTimestamp != false {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", false, result.Medias[0].IsFallbackedTimestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo3.jpg")
	if result.Medias[1].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[1].RelativePath)
	}
	if result.Medias[1].Timestamp != 1397367113 {
		t.Errorf("Expected timestamp %d, got %d", 1397367113, result.Medias[1].Timestamp)
	}
	if result.Medias[1].IsFallbackedTimestamp != false {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", false, result.Medias[1].IsFallbackedTimestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo1.jpg")
	if result.Medias[2].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[2].RelativePath)
	}
	if result.Medias[2].Timestamp != 1397367114 {
		t.Errorf("Expected timestamp %d, got %d", 1397367114, result.Medias[2].Timestamp)
	}
	if result.Medias[2].IsFallbackedTimestamp != false {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", false, result.Medias[2].IsFallbackedTimestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo4.jpg")
	if result.Medias[3].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[3].RelativePath)
	}
	if result.Medias[3].Timestamp != 1397367115 {
		t.Errorf("Expected timestamp %d, got %d", 1397367115, result.Medias[3].Timestamp)
	}
	if result.Medias[3].IsFallbackedTimestamp != false {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", false, result.Medias[3].IsFallbackedTimestamp)
	}

	wantPath = filepath.Join(temp, "dir2", "Takeout", "Google フォト", "2014", "photo5")
	if result.Medias[4].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[4].RelativePath)
	}
	if result.Medias[4].Timestamp != 1397367116 {
		t.Errorf("Expected timestamp %d, got %d", 13973671156, result.Medias[4].Timestamp)
	}
	if result.Medias[4].IsFallbackedTimestamp != true {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", true, result.Medias[4].IsFallbackedTimestamp)
	}
	wantPath = filepath.Join(temp, "dir2", "Takeout", "Google フォト", "2014", "photo5.jpg")
	if result.Medias[5].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[5].RelativePath)
	}
	if result.Medias[5].Timestamp != 1397367116 {
		t.Errorf("Expected timestamp %d, got %d", 1397367116, result.Medias[5].Timestamp)
	}
	if result.Medias[5].IsFallbackedTimestamp != false {
		t.Errorf("Expected IsFallbackedTimestamp %v, got %v", false, result.Medias[5].IsFallbackedTimestamp)
	}

	// Check result.Unexpected
	wantPath = filepath.Join("Takeout", "Google フォト", "2014", "photo3.jpg")
	if len(result.Unexpected.DuplicatedMetadata) != 1 || result.Unexpected.DuplicatedMetadata[0] != wantPath {
		t.Errorf("Expected duplicated metadata for %s, but %v", wantPath, result.Unexpected.DuplicatedMetadata)
	}
	wantPath = filepath.Join("Takeout", "Google フォト", "2014", "photo4.jpg")
	if len(result.Unexpected.DuplicatedMedia) != 1 || result.Unexpected.DuplicatedMedia[0] != wantPath {
		t.Errorf("Expected duplicated media file for %s, but %v", wantPath, result.Unexpected.DuplicatedMedia)
	}
	wantPath = filepath.Join("Takeout", "Google フォト", "2014", "no_metadata.jpg")
	if len(result.Unexpected.NotFoundMetadata) != 1 || result.Unexpected.NotFoundMetadata[0] != wantPath {
		t.Errorf("Expected not found metadata file for %s, but %v", wantPath, result.Unexpected.NotFoundMetadata)
	}
	wantPath = filepath.Join("Takeout", "Google フォト", "2014", "no_media.jpg")
	if len(result.Unexpected.NotFoundMedia) != 1 || result.Unexpected.NotFoundMedia[0] != wantPath {
		t.Errorf("Expected not found media file for %s, but %v", wantPath, result.Unexpected.NotFoundMedia)
	}
}

func TestTakeoutMetadataRepo_AnalyzeAllMetadata_inValidJson(t *testing.T) {
	temp := t.TempDir()
	files := map[string][]byte{
		// media and metadata in dir1
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.jpg"):  []byte(""),
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.json"): append(makeJson("photo1.jpg", "1397367114"), []byte(" invalid")...),
	}
	for path, content := range files {
		fullPath := filepath.Join(temp, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	repo := NewTakeoutMetadataRepo()
	result, err := repo.AnalyzeAllMetadata(temp)

	if err == nil {
		t.Fatalf("AnalyzeAllMetadata returned non error, expected json error: %v", result)
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("Expected error type json.SyntaxError: %v, %T", err, err)
	}
}

func TestTakeoutMetadataRepo_AnalyzeAllMetadata_inValidTypeJson(t *testing.T) {
	temp := t.TempDir()
	files := map[string][]byte{
		// media and metadata in dir1
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.jpg"):  []byte(""),
		filepath.Join("dir1", "Takeout", "Google フォト", "2014", "photo1.json"): []byte("{\"title\":\"photo1.jpg\",\"photoTakenTime\":{\"timestamp\":1}}"),
	}
	for path, content := range files {
		fullPath := filepath.Join(temp, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	repo := NewTakeoutMetadataRepo()
	result, err := repo.AnalyzeAllMetadata(temp)

	if err == nil {
		t.Fatalf("AnalyzeAllMetadata returned non error, expected json error: %v", result)
	}
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("Expected error type json.UnmarshalTypeError: %v, %T", err, err)
	}
}

func TestTakeoutMetadataRepo_AnalyzeAllMetadata_notFoundTakeoutDirectory(t *testing.T) {
	temp := t.TempDir()
	files := map[string][]byte{
		// media and metadata in dir1
		filepath.Join("dir1", "Takeouts", "Google フォト", "2014", "photo1.jpg"):  []byte(""),
	}
	for path, content := range files {
		fullPath := filepath.Join(temp, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	repo := NewTakeoutMetadataRepo()
	result, err := repo.AnalyzeAllMetadata(temp)

	if err == nil {
		t.Fatalf("AnalyzeAllMetadata returned non error, expected an error: %v", result)
	}
	message := err.Error()
	wantMessage := fmt.Sprintf("The path (%s) doesn't have Takeout.", filepath.Join(temp, "dir1", "Takeouts", "Google フォト", "2014", "photo1.jpg"))
	if message != wantMessage {
		t.Fatalf("Expected error message `%s`, got `%s`", wantMessage, message)
	}
}

func makeJson(title string, timestamp string) []byte {
	m := metadataJson{Title: title}
	m.PhotoTakenTime.Timestamp = timestamp
	b, _ := json.Marshal(m)
	return b
}
