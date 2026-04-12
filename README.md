## g-takeout-porter
This program organizes Google Phots data exported via Google Takeout into monthly directories with a specified location.

## How to use
```
$ go build ./cmd/takeout-porter/
$ ./takeout-porter source-directory dest-directory
```

## Behavior
### Input
You should unziped takeout zip files in source directory.
```
source-directroy
 +- some-directories1/Takeout/some-directories/photo1.jpg
 |                                            +photo1.jpg.json
 |                                            +photo2.jpg.json
 +- some-directories2/Takeout/some-directories/photo2.jpg
 .
 .
```
### Output
In dest-directory
- Creates monthly directories (e.g., dest-directory/202604/) under the specified path and organizes files into them.
- Updates the modification date of each copied file based on the metadata included in the export.

## TODO
- Check exif
- Check exisiting file in dest directory
- Add option
  - `strict mode`: disable fallback timestamp