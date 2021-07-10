package iso9660

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regresstion test for github.com/kdomanski/iso9660/issues/9
func TestRegressionNameWithoutDot(t *testing.T) {
	var buf bytes.Buffer

	//
	// create image with a file without dot
	//
	w, err := NewWriter()
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := w.Cleanup(); cleanupErr != nil {
			t.Fatalf("failed to cleanup writer: %v", cleanupErr)
		}
	}()

	err = w.AddFile(strings.NewReader("hrh2309hr320h"), "NODOT")
	assert.NoError(t, err)

	err = w.WriteTo(&buf, "testvolume")
	assert.NoError(t, err)

	//
	// no read it
	//

	image, err := OpenImage(bytes.NewReader(buf.Bytes()))
	assert.NoError(t, err)

	rootDir, err := image.RootDir()
	assert.NoError(t, err)

	children, err := rootDir.GetChildren()
	assert.NoError(t, err)

	nodotfile := children[0]
	assert.Equal(t, "nodot", nodotfile.Name())
}
