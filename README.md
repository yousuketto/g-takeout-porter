## g-takeout-porter
This program organizes Google Phots data exported via Google Takeout into monthly directories with a specified location.

## How to use
```
$ go build ./cmd/takeout-porter/
$ ./takeout-porter source-directory dest-directory
```

## Directory structure after execution
- Creates monthly directories (e.g., dest-directory/202604/) under the specified path and organizes files into them.
- Updates the modification date of each copied file based on the metadata included in the export.

## TODO
- Check exif
- Check exisiting file in dest directory
- Add option
  - `strict mode`: disable fallback timestamp