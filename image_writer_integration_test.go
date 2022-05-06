//go:build integration
// +build integration

package iso9660

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Identical to TestWriterAndMountWithAddLocal, except it uses
// AddFile to add files with a constant string in them.
func TestWriterAndMount(t *testing.T) {
	//
	// Image creation
	//

	// create a new image writer and prepare for cleanup
	w, err := NewWriter()
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := w.Cleanup(); cleanupErr != nil {
			t.Fatalf("failed to cleanup writer: %v", cleanupErr)
		}
	}()

	// add 1000 files to thoroughly test descriptor writing over sector bounds
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("firstlevel/dir%d/thirdlevel/file%d.txt", i, i)
		err = w.AddFile(strings.NewReader("hrh2309hr320h"), path)
		assert.NoError(t, err)
	}

	// write test image
	f, err := ioutil.TempFile(os.TempDir(), "iso9660_golang_test")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	err = w.WriteTo(f, "testvolume")
	assert.NoError(t, err)

	//
	// Image mounting
	//

	// create mount directory
	mountDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer func() {
		if removeErr := os.RemoveAll(mountDir); removeErr != nil {
			t.Fatalf("failed to delete mount directory: %v", removeErr)
		}
	}()

	// execute mount
	mountCmd := exec.Command("mount", "-t", "iso9660", f.Name(), mountDir)
	output, err := mountCmd.CombinedOutput()
	assert.NoError(t, err, "failed to mount the ISO image: %v\n%s", err, string(output))

	// make sure that we can see the 1000 files
	txtFileCount := 0
	walkfn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if string(data) != "hrh2309hr320h" {
			return fmt.Errorf("file %q has the wrong contents", path)
		}

		txtFileCount++
		return nil
	}
	err = filepath.Walk(mountDir, walkfn)
	assert.NoError(t, err)
	assert.Equal(t, 1000, txtFileCount)

	// execute umount
	umountCmd := exec.Command("umount", mountDir)
	output, err = umountCmd.CombinedOutput()
	assert.NoError(t, err, "failed to unmount the ISO image: %v\n%s", err, string(output))
}

// Identical to TestWriterAndMount, except it uses
// AddLocalFile to add an existing file.
func TestWriterAndMountWithAddLocal(t *testing.T) {
	sampleData, err := ioutil.ReadFile("/etc/issue")
	assert.NoError(t, err)

	//
	// Image creation
	//

	// create a new image writer and prepare for cleanup
	w, err := NewWriter()
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := w.Cleanup(); cleanupErr != nil {
			t.Fatalf("failed to cleanup writer: %v", cleanupErr)
		}
	}()

	// add 1000 files to thoroughly test descriptor writing over sector bounds
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("firstlevel/dir%d/thirdlevel/file%d.txt", i, i)
		err = w.AddLocalFile("/etc/issue", path)
		assert.NoError(t, err)
	}

	// write test image
	f, err := ioutil.TempFile(os.TempDir(), "iso9660_golang_test")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	err = w.WriteTo(f, "testvolume")
	assert.NoError(t, err)

	//
	// Image mounting
	//

	// create mount directory
	mountDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer func() {
		if removeErr := os.RemoveAll(mountDir); removeErr != nil {
			t.Fatalf("failed to delete mount directory: %v", removeErr)
		}
	}()

	// execute mount
	mountCmd := exec.Command("mount", "-t", "iso9660", f.Name(), mountDir)
	output, err := mountCmd.CombinedOutput()
	assert.NoError(t, err, "failed to mount the ISO image: %v\n%s", err, string(output))

	// make sure that we can see the 1000 files
	txtFileCount := 0
	walkfn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if string(data) != string(sampleData) {
			return fmt.Errorf("file %q has the wrong contents", path)
		}

		txtFileCount++
		return nil
	}
	err = filepath.Walk(mountDir, walkfn)
	assert.NoError(t, err)
	assert.Equal(t, 1000, txtFileCount)

	// execute umount
	umountCmd := exec.Command("umount", mountDir)
	output, err = umountCmd.CombinedOutput()
	assert.NoError(t, err, "failed to unmount the ISO image: %v\n%s", err, string(output))
}
