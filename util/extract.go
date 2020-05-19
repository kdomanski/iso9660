package util

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/kdomanski/iso9660"
)

func ExtractImageToDirectory(image io.ReaderAt, destination string) error {
	img, err := iso9660.OpenImage(image)
	if err != nil {
		return err
	}

	root, err := img.RootDir()
	if err != nil {
		return err
	}

	return extract(root, destination)

}

func extract(f *iso9660.File, targetPath string) error {
	// if f.Name() != string([]byte{0}) {
	// 	targetPath = path.Join(targetPath, f.Name())
	// }

	if f.IsDir() {
		existing, err := os.Open(targetPath)
		if err == nil {
			defer existing.Close()
			s, err := existing.Stat()
			if err != nil {
				return err
			}

			if !s.IsDir() {
				return fmt.Errorf("%s already exists and is a file", targetPath)
			}
		} else if os.IsNotExist(err) {
			if err = os.Mkdir(targetPath, 0755); err != nil {
				return err
			}
		} else {
			return err
		}

		children, err := f.GetChildren()
		if err != nil {
			return err
		}

		for _, c := range children {
			if err = extract(c, path.Join(targetPath, c.Name())); err != nil {
				return err
			}
		}
	} else { // it's a file
		newFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer newFile.Close()
		if _, err = io.Copy(newFile, f.Reader()); err != nil {
			return err
		}
	}

	return nil
}
