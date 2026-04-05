package infra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
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
			name:      "Normal: The path contains one 'Takeout' directory",
			input:     filepath.Join("some", "dir", "Takeout", "Google フォト", "photo.jpg"),
			want:      []string{"Takeout", "Google フォト", "photo.jpg"},
			wantError: false,
		},
		{
			name:      "Normal: The path contains two or more 'Takeout' directories",
			input:     filepath.Join("first", "Takeout", "second", "Takeout", "directory", "photo.img"),
			want:      []string{"Takeout", "directory", "photo.img"},
			wantError: false,
		},
		{
			name:      "Abnormal: The path doesn't contain a 'Takeout' directory",
			input:     filepath.Join("some", "dir", "Takeouts", "Google フォト", "photo.jpg"),
			want:      nil,
			wantError: true,
		},
		{
			name:      "Abnormal: The path doesn't contain a 'Takeout' directory, but contains a 'Takeout' file",
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
	if len(result.Medias) != 4 {
		t.Fatalf("AnalyzeAllMetadata returned %d medias, expected %d", len(result.Medias), 4)
		return
	}
	sort.Slice(result.Medias, func(i, j int) bool {
		return result.Medias[i].Timestamp < result.Medias[j].Timestamp
	})
	// Check result.Medias
	wantPath := filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo2.jpg")
	if result.Medias[0].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[0].RelativePath)
	}
	if result.Medias[0].Timestamp != 1397367111 {
		t.Errorf("Expected timestamp %d, got %d", 1397367111, result.Medias[0].Timestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo3.jpg")
	if result.Medias[1].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[0].RelativePath)
	}
	if result.Medias[1].Timestamp != 1397367113 {
		t.Errorf("Expected timestamp %d, got %d", 1397367113, result.Medias[0].Timestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo1.jpg")
	if result.Medias[2].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[0].RelativePath)
	}
	if result.Medias[2].Timestamp != 1397367114 {
		t.Errorf("Expected timestamp %d, got %d", 1397367114, result.Medias[0].Timestamp)
	}
	wantPath = filepath.Join(temp, "dir1", "Takeout", "Google フォト", "2014", "photo4.jpg")
	if result.Medias[3].RelativePath != wantPath {
		t.Errorf("Expected RelativePath %s, got %s", wantPath, result.Medias[0].RelativePath)
	}
	if result.Medias[3].Timestamp != 1397367115 {
		t.Errorf("Expected timestamp %d, got %d", 1397367115, result.Medias[0].Timestamp)
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

func makeJson(title string, timestamp string) []byte {
	m := metadataJson{Title: title}
	m.PhotoTakenTime.Timestamp = timestamp
	b, _ := json.Marshal(m)
	return b
}
