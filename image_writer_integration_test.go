//go:build integration
// +build integration

package iso9660

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func osDependentMountCmd(t *testing.T, filename string) (string, func()) {
	t.Helper()
	// Mount-DiskImage -ImagePath "E:\ISO-Files\My US Visit Fall 2010 Pictures.iso"
	switch runtime.GOOS {
	case "windows":
		ps, err := exec.LookPath("powershell.exe")
		assert.NoError(t, err)

		cmd := exec.Command(ps, "-NonInteractive", fmt.Sprintf("Mount-DiskImage -Verbose -ImagePath %q", filename))
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "failed to mount the ISO image: %v\n%s", err, string(output))

		cleanup := func() {
			// execute umount
			umountCmd := exec.Command("Dismount-DiskImage", "-ImagePath", filename)
			output, err = umountCmd.CombinedOutput()
			assert.NoError(t, err, "failed to unmount the ISO image: %v\n%s", err, string(output))
		}

		// We're assuming that the assigned letter will be D:
		return `D:\\`, cleanup
	case "linux":
		// create mount directory
		mountDir, err := ioutil.TempDir("", "")
		assert.NoError(t, err)

		cmd := exec.Command("mount", "-t", "iso9660", filename, mountDir)
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "failed to mount the ISO image: %v\n%s", err, string(output))

		cleanup := func() {
			// execute umount
			umountCmd := exec.Command("umount", mountDir)
			output, err = umountCmd.CombinedOutput()
			assert.NoError(t, err, "failed to unmount the ISO image: %v\n%s", err, string(output))

			if removeErr := os.RemoveAll(mountDir); removeErr != nil {
				t.Fatalf("failed to delete mount directory: %v", removeErr)
			}
		}

		return mountDir, cleanup
	}

	t.Fatalf("unknown OS %s", runtime.GOOS)
	return "", nil
}

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

	mountDir, cleanup := osDependentMountCmd(t, f.Name())
	defer cleanup()

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

}

// Identical to TestWriterAndMount, except it uses
// AddLocalFile to add an existing file.
func TestWriterAndMountWithAddLocal(t *testing.T) {
	var testFile string

	switch runtime.GOOS {
	case "windows":
		testFile = `C:\Windows\System32\notepad.exe`
	case "linux":
		testFile = "/etc/issue"
	default:
		t.Fatalf("unknown OS %s", runtime.GOOS)
	}

	sampleData, err := ioutil.ReadFile(testFile)
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
		err = w.AddLocalFile(testFile, path)
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

	// execute mount
	mountDir, cleanup := osDependentMountCmd(t, f.Name())
	defer cleanup()

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
}
