## g-takeout-porter
A CLI tool to organize Google Photos data exported via Google Takeout into monthly directories based on metadata.

## Usage
```bash
$ go build ./cmd/takeout-porter/
$ ./takeout-porter <source-directory> <dest-directory>
```

## Google Takeout Settings (Important)
To ensure this tool works correctly and to avoid duplicate files, please follow these settings when creating your export:

1.  **Select only "Google Photos"**: Deselect all other products.
2.  **Filter Albums**: Click "All photo albums included" and **deselect all custom albums**.
    *   Only keep the folders named by year (e.g., "Photos from 2024").
    *   Including custom albums will result in duplicate files in the export, which may complicate the organization process.

## How It Works
### Input
Unzip your Google Takeout archives into a source directory. The tool scans for media files and their corresponding `.json` metadata.

```
source-directory
 +- folder1/Takeout/Google Photos/Photos from 2024/photo1.jpg
 |                                                +photo1.jpg.json
 |                                                +photo2.jpg.json
 +- folder2/Takeout/Google Photos/Photos from 2024/photo2.jpg
 ...
```

### Output
In the destination directory, the tool creates monthly directories and organizes files into them.

```
dest-directory
 +- 202401/photo1.jpg
 |        +photo2.jpg
 +- 202402/photo3.jpg
 +- 202403/photo4.jpg
 ...
```

- **Monthly Organization**: Files are grouped into folders named `YYYYMM` (e.g., `202401`).
- **Timestamp Restoration**: Updates the modification date (`mtime`) of each copied file using the timestamp from the Takeout metadata.

## TODO
- [ ] Support EXIF metadata extraction
- [ ] Check for existing files in the destination directory to avoid duplicates
- [ ] Add CLI options:
  - `--strict`: Disable fallback timestamp logic
