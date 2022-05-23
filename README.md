## iso9660
[![GoDoc](https://godoc.org/github.com/kdomanski/iso9660?status.svg)](https://godoc.org/github.com/kdomanski/iso9660)
[![codecov](https://codecov.io/gh/kdomanski/iso9660/branch/master/graph/badge.svg?token=14MJSZYZ24)](https://codecov.io/gh/kdomanski/iso9660)

A package for reading and creating ISO9660

Joliet and Rock Ridge extensions are not supported.

## Examples

### Extracting an ISO

```go
package main

import (
  "log"

  "github.com/kdomanski/iso9660/util"
)

func main() {
  f, err := os.Open("/home/user/myImage.iso")
  if err != nil {
    log.Fatalf("failed to open file: %s", err)
  }
  defer f.Close()

  if err = util.ExtractImageToDirectory(f, "/home/user/target_dir"); err != nil {
    log.Fatalf("failed to extract image: %s", err)
  }
}
```

### Creating an ISO

```go
package main

import (
  "log"
  "os"

  "github.com/kdomanski/iso9660"
)

func main() {
  writer, err := iso9660.NewWriter()
  if err != nil {
    log.Fatalf("failed to create writer: %s", err)
  }
  defer writer.Cleanup()

  f, err := os.Open("/home/user/myFile.txt")
  if err != nil {
    log.Fatalf("failed to open file: %s", err)
  }
  defer f.Close()

  err = writer.AddFile(f, "folder/MYFILE.TXT")
  if err != nil {
    log.Fatalf("failed to add file: %s", err)
  }

  outputFile, err := os.OpenFile("/home/user/output.iso", os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0644)
  if err != nil {
    log.Fatalf("failed to create file: %s", err)
  }

  err = writer.WriteTo(outputFile, "testvol")
  if err != nil {
    log.Fatalf("failed to write ISO image: %s", err)
  }

  err = outputFile.Close()
  if err != nil {
    log.Fatalf("failed to close output file: %s", err)
  }
}
```

### Recursively create an ISO image from the given directories

```go
package main

import (
  "fmt"
  "log"
  "os"
  "path/filepath"
  "strings"

  "github.com/kdomanski/iso9660"
)

func main() {
  writer, err := iso9660.NewWriter()
  if err != nil {
    log.Fatalf("failed to create writer: %s", err)
  }
  defer writer.Cleanup()

  isoFile, err := os.OpenFile("C:/output.iso", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
  if err != nil {
    log.Fatalf("failed to create file: %s", err)
  }
  defer isoFile.Close()

  prefix := "F:\\" // the prefix to remove in the output iso file
  sourceFolders := []string{"F:\\test1", "F:\\test2"} // the given directories to create an ISO file from

  for _, folderName := range sourceFolders {
    folderPath := strings.Join([]string{prefix, folderName}, "/")

    walk_err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
      if err != nil {
        log.Fatalf("walk: %s", err)
        return err
      }
      if info.IsDir() {
        return nil
      }
      outputPath := strings.TrimPrefix(path, prefix) // remove the source drive name
      fmt.Printf("Adding file: %s\n", outputPath)

      fileToAdd, err := os.Open(path)
      if err != nil {
        log.Fatalf("failed to open file: %s", err)
      }
      defer fileToAdd.Close()

      err = writer.AddFile(fileToAdd, outputPath)
      if err != nil {
        log.Fatalf("failed to add file: %s", err)
      }
      return nil
    })
    if walk_err != nil {
      log.Fatalf("%s", walk_err)
    }
  }

  err = writer.WriteTo(isoFile, "Test")
  if err != nil {
    log.Fatalf("failed to write ISO image: %s", err)
  }
}
```
