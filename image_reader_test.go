// +build !integration

package iso9660

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
`

func TestImageReader(t *testing.T) {
	tz := time.FixedZone("", 3600*2)
	recordTime := time.Date(2018, 07, 25, 22, 01, 02, 0, tz)

	f, err := os.Open("fixtures/test.iso")
	assert.NoError(t, err)
	defer f.Close() // nolint: errcheck

	image, err := OpenImage(f)
	if assert.NoError(t, err) {
		assert.Equal(t, 2, len(image.volumeDescriptors))
		assert.Equal(t, volumeTypePrimary, image.volumeDescriptors[0].Header.Type)
		assert.Equal(t, volumeTypeTerminator, image.volumeDescriptors[1].Header.Type)
	}

	rootDir, err := image.RootDir()
	assert.NoError(t, err)
	assert.True(t, rootDir.IsDir())
	assert.Equal(t, string([]byte{0}), rootDir.Name())
	assert.Equal(t, int64(sectorSize), rootDir.Size())
	assert.Equal(t, recordTime, rootDir.ModTime())

	children, err := rootDir.GetChildren()
	assert.NoError(t, err)
	assert.Len(t, children, 4)

	cicero := children[0]
	dir1 := children[1]
	dir2 := children[2]
	dir4 := children[3]

	assert.Equal(t, "CICERO.TXT", cicero.Name())
	assert.Equal(t, int64(845), cicero.Size())

	if assert.Equal(t, "DIR1", dir1.Name()) {
		dir1Children, err := dir1.GetChildren()
		assert.NoError(t, err)
		assert.Len(t, dir1Children, 1)

		loremFile := dir1Children[0]

		assert.Equal(t, "LOREM_IP.TXT", loremFile.Name())
		assert.Equal(t, int64(446), loremFile.Size())

		data, err := ioutil.ReadAll(loremFile.Reader())
		assert.NoError(t, err)

		assert.Equal(t, loremIpsum, string(data))
	}

	if assert.Equal(t, "DIR2", dir2.Name()) {
		dir2Children, err := dir2.GetChildren()
		assert.NoError(t, err)
		assert.Len(t, dir2Children, 2)

		assert.Equal(t, "DIR3", dir2Children[0].Name())
		dir3Children, err := dir2Children[0].GetChildren()
		assert.NoError(t, err)
		assert.Len(t, dir3Children, 1)

		assert.Equal(t, "DATA.BIN", dir3Children[0].Name())
		assert.Equal(t, int64(512), dir3Children[0].Size())

		assert.Equal(t, "LARGE.TXT", dir2Children[1].Name())
		assert.Equal(t, int64(2808), dir2Children[1].Size())
		assert.False(t, dir2Children[1].IsDir())
	}

	if assert.Equal(t, "DIR4", dir4.Name()) {
		dir4Children, err := dir4.GetChildren()
		assert.NoError(t, err)
		assert.Len(t, dir4Children, 1000)
		assert.Equal(t, "FILE1012", dir4Children[12].Name())
	}
}
