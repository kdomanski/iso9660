//go:build integration_create
// +build integration_create

package iso9660

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterCreate(t *testing.T) {
	writerTestImagePath := os.Getenv("ISO_WRITER_TEST_IMAGE")
	if !assert.NotEmpty(t, writerTestImagePath) {
		t.FailNow()
	}

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
	f, err := os.Create(writerTestImagePath)
	assert.NoError(t, err)

	err = w.WriteTo(f, "testvolume")
	assert.NoError(t, err)
}
