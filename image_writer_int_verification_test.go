//go:build integration_verify
// +build integration_verify

package iso9660

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Identical to TestWriterAndMountWithAddLocal, except it uses
// AddFile to add files with a constant string in them.
func TestWriterVerify(t *testing.T) {
	writerTestImagePath := os.Getenv("ISO_WRITER_TEST_IMAGE")
	if !assert.NotEmpty(t, writerTestImagePath) {
		t.FailNow()
	}

	//
	// Image mounting
	//

	// create mount directory
	mountDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)
	defer func() {
		if removeErr := os.RemoveAll(mountDir); removeErr != nil {
			t.Fatalf("failed to delete mount directory: %v", removeErr)
		}
	}()

	// execute mount
	mountCmd := exec.Command("mount", "-t", "iso9660", writerTestImagePath, mountDir)
	output, err := mountCmd.CombinedOutput()
	assert.NoError(t, err, "failed to mount the ISO image: %v\n%s", err, string(output))

	// make sure that we can see the 1000 files
	txtFileCount := 0
	walkfn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
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
