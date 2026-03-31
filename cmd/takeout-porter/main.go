package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

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

type FileMetadataJson struct {
	Title          string `json:"title"`
	PhotoTakenTime struct {
		Timestamp string `json:timestamp`
	} `json:photoTakenTime`
}

type SourceFileInfo struct {
	RelativePath string
	Timestamp    int64
	Key          string
}

func analyzeSourceDir(sourceDir string) ([]SourceFileInfo, error) {
	timestampMap := make(map[string]int64)
	pathMap := make(map[string]string)
	err := filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var m FileMetadataJson
			if err := json.Unmarshal(data, &m); err != nil {
				return fmt.Errorf("Fail to read json (%s): %v", path, err)
			}
			if m.Title == "" {
				return fmt.Errorf("Title is not found in %s", path)
			}

			takeoutRelativePath, err := extractTakeoutRelativePath(path)
			if err != nil {
				return err
			}

			ts, err := strconv.ParseInt(m.PhotoTakenTime.Timestamp, 10, 64)
			if err != nil {
				return fmt.Errorf("Fail to conver %s timestamp value: %v", m.Title, err)
			}

			key := filepath.Join(filepath.Join(takeoutRelativePath[:len(takeoutRelativePath)-1]...), m.Title)
			if _, existing := timestampMap[key]; existing {
				return fmt.Errorf("Expected file names are duplicated. Check metadata JSON files. (%s in %s)", key, path)
			}
			timestampMap[key] = ts
		} else {
			takeoutRelativePath, err := extractTakeoutRelativePath(path)
			if err != nil {
				return err
			}

			key := filepath.Join(takeoutRelativePath...)
			if _, existing := pathMap[key]; existing {
				return fmt.Errorf("File names are duplicated. Check '%s' in unziped directories.", key)
			}
			pathMap[key] = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sourceFileInfos := make([]SourceFileInfo, 0, len(pathMap))
	for k, relativePath := range pathMap {
		if timestamp, ok := timestampMap[k]; ok {
			info := SourceFileInfo{relativePath, timestamp, k}
			sourceFileInfos = append(sourceFileInfos, info)
			delete(timestampMap, k)
		} else {
			return nil, fmt.Errorf("Not found metadata file for '%s'", relativePath)
		}
	}
	if len(timestampMap) > 0 {
		return nil, fmt.Errorf("Not found real files for '%s'", strings.Join(slices.Collect(maps.Keys(timestampMap)), ", "))
	}

	return sourceFileInfos, nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}

func copyToDestDir(sourceFileInfos []SourceFileInfo, destDir string) error {
	for _, info := range sourceFileInfos {
		sourcePath := info.RelativePath

		timestamp := time.Unix(info.Timestamp, 0)

		targetPath := filepath.Join(destDir, timestamp.Format("200601"), filepath.Base(info.RelativePath))
		os.MkdirAll(filepath.Dir(targetPath), 0755)

		err := copyFile(sourcePath, targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s [source directory] [dest directory]\n", os.Args[0])
		os.Exit(1)
	}
	sourceDir := filepath.Clean(os.Args[1])
	destDir := filepath.Clean(os.Args[2])

	sourceFileInfos, err := analyzeSourceDir(sourceDir)
	if err != nil {
		log.Fatalf("Fail to analyze data structure in source directory: %v", err)
	}

	if err := copyToDestDir(sourceFileInfos, destDir); err != nil {
		log.Fatalf("Fail to copy files: %v", err)
	}

	for _, info := range sourceFileInfos {
		fmt.Printf("> %v\n", info)
	}

}
